package engine

import (
	"fmt"
	"strconv"
	"strings"
	model "tragedylooper/internal/game/proto/v1"
)

func (ge *GameEngine) applyEffect(effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	choices, err := ge.resolveEffectChoices(ge.GameState, effect, payload)
	if err != nil {
		return fmt.Errorf("error resolving choices: %w", err)
	}

	if len(choices) > 0 && choice == nil {
		choiceEvent := &model.ChoiceRequiredEvent{Choices: choices}
		ge.applyAndPublishEvent(model.GameEventType_CHOICE_REQUIRED, choiceEvent)
		return nil
	}

	err = ge.publishEffect(ge.GameState, effect, ability, payload, choice)
	if err != nil {
		return fmt.Errorf("error publishing effect: %w", err)
	}

	return nil
}

func (ge *GameEngine) resolveEffectChoices(state *model.GameState, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	switch t := effect.EffectType.(type) {
	case *model.Effect_CompoundEffect:
		return ge.resolveCompoundEffectChoices(state, t.CompoundEffect, payload)
	case *model.Effect_AdjustStat:
		return ge.createChoicesFromSelector(state, t.AdjustStat.Target, payload, "Select character to adjust stat")
	case *model.Effect_MoveCharacter:
		return ge.createChoicesFromSelector(state, t.MoveCharacter.Target, payload, "Select character to move")
	case *model.Effect_AddTrait:
		return ge.createChoicesFromSelector(state, t.AddTrait.Target, payload, "Select character to add trait to")
	case *model.Effect_RemoveTrait:
		return ge.createChoicesFromSelector(state, t.RemoveTrait.Target, payload, "Select character to remove trait from")
	}
	return nil, nil
}

func (ge *GameEngine) resolveCompoundEffectChoices(state *model.GameState, effect *model.CompoundEffect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	switch effect.Operator {
	case model.CompoundEffect_CHOOSE_ONE:
		var choices []*model.Choice
		for i, subEffect := range effect.SubEffects {
			choiceID := fmt.Sprintf("effect_choice_%d", i)
			choices = append(choices, &model.Choice{
				Id:          choiceID,
				Description: getEffectDescription(subEffect),
				ChoiceType:  &model.Choice_EffectOptionIndex{EffectOptionIndex: int32(i)},
			})
		}
		return choices, nil
	case model.CompoundEffect_SEQUENCE:
		for _, subEffect := range effect.SubEffects {
			choices, err := ge.resolveEffectChoices(state, subEffect, payload)
			if err != nil {
				return nil, err
			}
			if len(choices) > 0 {
				return choices, nil
			}
		}
	}
	return nil, nil
}

func (ge *GameEngine) createChoicesFromSelector(state *model.GameState, selector *model.TargetSelector, payload *model.UseAbilityPayload, description string) ([]*model.Choice, error) {
	// We pass a nil choice here because we are just trying to find out *if* a choice is needed.
	charIDs, err := ge.resolveSelectorToCharacters(state, selector, payload, nil)
	if err != nil {
		return nil, err
	}

	// If the selector resolves to more than one character, a choice is required.
	if len(charIDs) > 1 {
		var choices []*model.Choice
		for _, charID := range charIDs {
			char, ok := state.Characters[charID]
			if !ok {
				continue
			}
			choiceID := fmt.Sprintf("target_char_%d", charID)
			choices = append(choices, &model.Choice{
				Id:          choiceID,
				Description: fmt.Sprintf("%s: %s", description, char.Config.Name),
				ChoiceType:  &model.Choice_TargetCharacterId{TargetCharacterId: charID},
			})
		}
		return choices, nil
	}

	return nil, nil
}

func (ge *GameEngine) resolveSelectorToCharacters(state *model.GameState, selector *model.TargetSelector, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) ([]int32, error) {
	// 1. Check if a choice was already made and provided.
	if choice != nil {
		choiceID := choice.GetChosenOptionId()
		if strings.HasPrefix(choiceID, "target_char_") {
			idStr := strings.TrimPrefix(choiceID, "target_char_")
			id, err := strconv.ParseInt(idStr, 10, 32)
			if err == nil {
				return []int32{int32(id)}, nil
			}
		}
	}

	// 2. Check for a target in the initial payload (for abilities that directly target).
	if payload != nil && payload.GetTargetCharacterId() != 0 {
		return []int32{payload.GetTargetCharacterId()}, nil
	}

	// 3. Resolve the selector based on its type.
	if handler, ok := selectorHandlers[selector.SelectorType]; ok {
		return handler(ge, state, selector, payload)
	}

	// Fallback to all characters if selector is not specific and no target is provided
	var allCharIDs []int32
	for id := range state.Characters {
		allCharIDs = append(allCharIDs, id)
	}
	return allCharIDs, nil
}

type selectorHandler func(ge *GameEngine, state *model.GameState, selector *model.TargetSelector, payload *model.UseAbilityPayload) ([]int32, error)

var selectorHandlers = map[model.TargetSelector_SelectorType]selectorHandler{
	model.TargetSelector_SPECIFIC_CHARACTER:         resolveSpecificCharacter,
	model.TargetSelector_TRIGGERING_CHARACTER:       resolveTriggeringCharacter,
	model.TargetSelector_ALL_CHARACTERS_AT_LOCATION: resolveAllCharactersAtLocation,
	model.TargetSelector_ANY_CHARACTER_WITH_ROLE:    resolveAnyCharacterWithRole,
}

func resolveSpecificCharacter(ge *GameEngine, state *model.GameState, selector *model.TargetSelector, payload *model.UseAbilityPayload) ([]int32, error) {
	return []int32{selector.CharacterId}, nil
}

func resolveTriggeringCharacter(ge *GameEngine, state *model.GameState, selector *model.TargetSelector, payload *model.UseAbilityPayload) ([]int32, error) {
	if payload != nil {
		return []int32{payload.CharacterId}, nil
	}
	return nil, fmt.Errorf("could not resolve triggering character: payload is nil")
}

func resolveAllCharactersAtLocation(ge *GameEngine, state *model.GameState, selector *model.TargetSelector, payload *model.UseAbilityPayload) ([]int32, error) {
	var charIDs []int32
	for id, char := range state.Characters {
		if char.CurrentLocation == selector.LocationFilter {
			charIDs = append(charIDs, id)
		}
	}
	return charIDs, nil
}

func resolveAnyCharacterWithRole(ge *GameEngine, state *model.GameState, selector *model.TargetSelector, payload *model.UseAbilityPayload) ([]int32, error) {
	var charIDs []int32
	for id, char := range state.Characters {
		if char.HiddenRole == selector.RoleFilter {
			charIDs = append(charIDs, id)
		}
	}
	return charIDs, nil
}

func (ge *GameEngine) publishEffect(state *model.GameState, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	switch t := effect.EffectType.(type) {
	case *model.Effect_AdjustStat:
		return ge.publishAdjustStatEffect(state, t.AdjustStat, payload, choice)
	case *model.Effect_MoveCharacter:
		return ge.publishMoveCharacterEffect(state, t.MoveCharacter, payload, choice)
	case *model.Effect_AddTrait:
		return ge.publishAddTraitEffect(state, t.AddTrait, payload, choice)
	case *model.Effect_RemoveTrait:
		return ge.publishRemoveTraitEffect(state, t.RemoveTrait, payload, choice)
	case *model.Effect_CompoundEffect:
		return ge.publishCompoundEffect(state, t.CompoundEffect, ability, payload, choice)
	default:
		return fmt.Errorf("unknown effect type: %T", t)
	}
}

func (ge *GameEngine) publishAdjustStatEffect(state *model.GameState, effect *model.AdjustStatEffect, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, payload, choice)
	if err != nil {
		return err
	}

	for _, targetID := range targetIDs {
		char, ok := state.Characters[targetID]
		if !ok {
			continue // Or return an error
		}

		switch effect.StatType {
		case model.StatCondition_PARANOIA:
			newParanoia := char.Paranoia + effect.Amount
			event := &model.ParanoiaAdjustedEvent{CharacterId: targetID, NewParanoia: newParanoia, Amount: effect.Amount}
			ge.applyAndPublishEvent(model.GameEventType_PARANOIA_ADJUSTED, event)
		case model.StatCondition_INTRIGUE:
			newIntrigue := char.Intrigue + effect.Amount
			event := &model.IntrigueAdjustedEvent{CharacterId: targetID, NewIntrigue: newIntrigue, Amount: effect.Amount}
			ge.applyAndPublishEvent(model.GameEventType_INTRIGUE_ADJUSTED, event)
		case model.StatCondition_GOODWILL:
			newGoodwill := char.Goodwill + effect.Amount
			event := &model.GoodwillAdjustedEvent{CharacterId: targetID, NewGoodwill: newGoodwill, Amount: effect.Amount}
			ge.applyAndPublishEvent(model.GameEventType_GOODWILL_ADJUSTED, event)
		}
	}
	return nil
}

func (ge *GameEngine) publishMoveCharacterEffect(state *model.GameState, effect *model.MoveCharacterEffect, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, payload, choice)
	if err != nil {
		return err
	}
	for _, targetID := range targetIDs {
		event := &model.CharacterMovedEvent{CharacterId: targetID, NewLocation: effect.Destination}
		ge.applyAndPublishEvent(model.GameEventType_CHARACTER_MOVED, event)
	}
	return nil
}

func (ge *GameEngine) publishAddTraitEffect(state *model.GameState, effect *model.AddTraitEffect, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, payload, choice)
	if err != nil {
		return err
	}
	for _, targetID := range targetIDs {
		event := &model.TraitAddedEvent{CharacterId: targetID, Trait: effect.Trait}
		ge.applyAndPublishEvent(model.GameEventType_TRAIT_ADDED, event)
	}
	return nil
}

func (ge *GameEngine) publishRemoveTraitEffect(state *model.GameState, effect *model.RemoveTraitEffect, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, payload, choice)
	if err != nil {
		return err
	}
	for _, targetID := range targetIDs {
		event := &model.TraitRemovedEvent{CharacterId: targetID, Trait: effect.Trait}
		ge.applyAndPublishEvent(model.GameEventType_TRAIT_REMOVED, event)
	}
	return nil
}

func (ge *GameEngine) publishCompoundEffect(state *model.GameState, effect *model.CompoundEffect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	switch effect.Operator {
	case model.CompoundEffect_SEQUENCE:
		for _, subEffect := range effect.SubEffects {
			// Pass the choice down; it might be needed by a later effect in the sequence
			err := ge.publishEffect(state, subEffect, ability, payload, choice)
			if err != nil {
				// If a sub-effect requires a choice and didn't get one, it will return an error.
				// We might need to check for a ChoiceRequiredEvent here.
				return err
			}
		}
	case model.CompoundEffect_CHOOSE_ONE:
		if choice == nil {
			// This should have been caught earlier, but as a safeguard:
			return fmt.Errorf("a choice is required to publish a CHOOSE_ONE compound effect")
		}
		choiceID := choice.GetChosenOptionId()
		if !strings.HasPrefix(choiceID, "effect_choice_") {
			return fmt.Errorf("invalid choice id for compound effect: %s", choiceID)
		}
		indexStr := strings.TrimPrefix(choiceID, "effect_choice_")
		choiceIndex, err := strconv.Atoi(indexStr)
		if err != nil {
			return fmt.Errorf("invalid choice index: %s", indexStr)
		}

		if choiceIndex < 0 || choiceIndex >= len(effect.SubEffects) {
			return fmt.Errorf("choice index out of bounds: %d", choiceIndex)
		}
		// Execute the chosen sub-effect
		err = ge.publishEffect(state, effect.SubEffects[choiceIndex], ability, payload, choice)
		if err != nil {
			return err
		}
	}
	return nil
}

// getEffectDescription provides a human-readable summary of an effect.
func getEffectDescription(effect *model.Effect) string {
	switch t := effect.EffectType.(type) {
	case *model.Effect_AdjustStat:
		return fmt.Sprintf("Adjust %s by %d", t.AdjustStat.StatType, t.AdjustStat.Amount)
	case *model.Effect_MoveCharacter:
		return fmt.Sprintf("Move character to %s", t.MoveCharacter.Destination)
	case *model.Effect_AddTrait:
		return fmt.Sprintf("Add trait '%s'", t.AddTrait.Trait)
	case *model.Effect_RemoveTrait:
		return fmt.Sprintf("Remove trait '%s'", t.RemoveTrait.Trait)
	case *model.Effect_CompoundEffect:
		return "Choose one of the following effects"
	default:
		return "(Unknown effect)"
	}
}
