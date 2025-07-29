package engine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	model "tragedylooper/pkg/proto/v1"
)

func (ge *GameEngine) applyEffect(effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	choices, err := ge.resolveEffectChoices(ge.GameState, effect, payload)
	if err != nil {
		return fmt.Errorf("error resolving choices: %w", err)
	}

	if len(choices) > 0 && choice == nil {
		choiceEvent := &model.ChoiceRequiredEvent{Choices: choices}
		ge.ApplyAndPublishEvent(model.GameEventType_CHOICE_REQUIRED, choiceEvent)
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
	charIDs, err := ge.resolveSelectorToCharacters(state, selector, nil, payload, nil)
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

func (ge *GameEngine) resolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, player *model.Player, payload *model.UseAbilityPayload, ability *model.Ability) ([]int32, error) {
	if sel == nil {
		return nil, errors.New("target selector is nil")
	}

	switch sel.SelectorType {
	case model.TargetSelector_ABILITY_USER:
		if ability == nil {
			return nil, errors.New("ability is nil for ABILITY_USER selector")
		}
		return []int32{ability.OwnerCharacterId}, nil
	case model.TargetSelector_ABILITY_TARGET:
		if payload == nil {
			return nil, errors.New("payload is nil for ABILITY_TARGET selector")
		}
		if targetChar, ok := payload.Target.(*model.UseAbilityPayload_TargetCharacterId); ok {
			return []int32{targetChar.TargetCharacterId}, nil
		}
		return nil, errors.New("payload does not contain a target character for ABILITY_TARGET selector")
	case model.TargetSelector_ALL_CHARACTERS:
		ids := make([]int32, 0, len(gs.Characters))
		for id := range gs.Characters {
			ids = append(ids, id)
		}
		return ids, nil
	case model.TargetSelector_SPECIFIC_CHARACTER:
		return []int32{sel.CharacterId}, nil
	// TODO: Implement other selector types
	default:
		return nil, errors.New("unsupported target selector type")
	}
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
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, nil, payload, nil)
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
			ge.ApplyAndPublishEvent(model.GameEventType_PARANOIA_ADJUSTED, event)
		case model.StatCondition_INTRIGUE:
			newIntrigue := char.Intrigue + effect.Amount
			event := &model.IntrigueAdjustedEvent{CharacterId: targetID, NewIntrigue: newIntrigue, Amount: effect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_INTRIGUE_ADJUSTED, event)
		case model.StatCondition_GOODWILL:
			newGoodwill := char.Goodwill + effect.Amount
			event := &model.GoodwillAdjustedEvent{CharacterId: targetID, NewGoodwill: newGoodwill, Amount: effect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_GOODWILL_ADJUSTED, event)
		}
	}
	return nil
}

func (ge *GameEngine) publishMoveCharacterEffect(state *model.GameState, effect *model.MoveCharacterEffect, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, nil, payload, nil)
	if err != nil {
		return err
	}
	for _, targetID := range targetIDs {
		char := ge.getCharacterByID(targetID)
		if char == nil {
			continue
		}
		ge.moveCharacter(char, 0, 0) // A generic move, let the moveCharacter logic handle the details
	}
	return nil
}

func (ge *GameEngine) publishAddTraitEffect(state *model.GameState, effect *model.AddTraitEffect, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, nil, payload, nil)
	if err != nil {
		return err
	}
	for _, targetID := range targetIDs {
		event := &model.TraitAddedEvent{CharacterId: targetID, Trait: effect.Trait}
		ge.ApplyAndPublishEvent(model.GameEventType_TRAIT_ADDED, event)
	}
	return nil
}

func (ge *GameEngine) publishRemoveTraitEffect(state *model.GameState, effect *model.RemoveTraitEffect, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, nil, payload, nil)
	if err != nil {
		return err
	}
	for _, targetID := range targetIDs {
		event := &model.TraitRemovedEvent{CharacterId: targetID, Trait: effect.Trait}
		ge.ApplyAndPublishEvent(model.GameEventType_TRAIT_REMOVED, event)
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
