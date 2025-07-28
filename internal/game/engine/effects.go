package engine

import (
	"fmt"
	"strconv"
	"strings"
	model "tragedylooper/internal/game/proto/v1"
)

func (ge *GameEngine) applyEffect(effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload) error {
	choices, err := ge.resolveEffectChoices(ge.GameState, effect, ability, payload)
	if err != nil {
		return fmt.Errorf("error resolving choices: %w", err)
	}

	if len(choices) > 0 && (payload == nil || payload.GetChooseOption() == nil || payload.GetChooseOption().GetChoiceId() == "") {
		choiceEvent := &model.ChoiceRequiredEvent{Choices: choices}
		ge.publishGameEvent(model.GameEventType_CHOICE_REQUIRED, choiceEvent)
		return nil
	}

	events, err := ge.executeEffect(ge.GameState, effect, ability, payload)
	if err != nil {
		return fmt.Errorf("error executing effect: %w", err)
	}

	for _, event := range events {
		ge.processEvent(event)
	}

	return nil
}

func (ge *GameEngine) resolveEffectChoices(state *model.GameState, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	switch t := effect.EffectType.(type) {
	case *model.Effect_CompoundEffect:
		switch t.CompoundEffect.Operator {
		case model.CompoundEffect_CHOOSE_ONE:
			var choices []*model.Choice
			for i, subEffect := range t.CompoundEffect.SubEffects {
				choiceID := fmt.Sprintf("effect_choice_%d", i)
				choices = append(choices, &model.Choice{
					Id:          choiceID,
					Description: fmt.Sprintf("Choose effect option %d", i), // Placeholder
					ChoiceType:  &model.Choice_EffectOptionIndex{EffectOptionIndex: int32(i)},
				})
			}
			return choices, nil
		case model.CompoundEffect_SEQUENCE:
			for _, subEffect := range t.CompoundEffect.SubEffects {
				choices, err := ge.resolveEffectChoices(state, subEffect, ability, payload)
				if err != nil {
					return nil, err
				}
				if len(choices) > 0 {
					return choices, nil
				}
			}
			return nil, nil
		}
	case *model.Effect_AdjustStat:
		return ge.createChoicesFromSelector(state, t.AdjustStat.Target, payload, "Select character to adjust stat")
	case *model.Effect_MoveCharacter:
		return ge.createChoicesFromSelector(state, t.MoveCharacter.Target, payload, "Select character to move")
	// ... other cases from previous implementation
	}
	return nil, nil
}

func (ge *GameEngine) createChoicesFromSelector(state *model.GameState, selector *model.TargetSelector, payload *model.UseAbilityPayload, description string) ([]*model.Choice, error) {
	charIDs, err := ge.resolveSelectorToCharacters(state, selector, nil) // Pass nil payload initially
	if err != nil {
		return nil, err
	}

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
				Description: fmt.Sprintf("%s: %s", description, char.Name),
				ChoiceType:  &model.Choice_TargetCharacterId{TargetCharacterId: charID},
			})
		}
		return choices, nil
	}

	return nil, nil
}

func (ge *GameEngine) resolveSelectorToCharacters(state *model.GameState, selector *model.TargetSelector, payload *model.UseAbilityPayload) ([]int32, error) {
	if payload != nil && payload.GetChooseOption() != nil {
		choiceID := payload.GetChooseOption().GetChoiceId()
		if strings.HasPrefix(choiceID, "target_char_") {
			idStr := strings.TrimPrefix(choiceID, "target_char_")
			id, err := strconv.Atoi(idStr)
			if err == nil {
				return []int32{int32(id)}, nil
			}
		}
	}
	
	if payload != nil && payload.GetTargetCharacterId() != 0 {
		return []int32{payload.GetTargetCharacterId()}, nil
	}

	switch selector.SelectorType {
	case model.TargetSelector_SPECIFIC_CHARACTER:
		return []int32{selector.CharacterId}, nil
	case model.TargetSelector_TRIGGERING_CHARACTER:
		if payload != nil && payload.GetTriggeringCharacterId() != 0 {
			return []int32{payload.GetTriggeringCharacterId()}, nil
		}
		return nil, fmt.Errorf("could not resolve triggering character")
	case model.TargetSelector_ALL_CHARACTERS_AT_LOCATION:
		var charIDs []int32
		for id, char := range state.Characters {
			if char.CurrentLocation == selector.LocationFilter {
				charIDs = append(charIDs, id)
			}
		}
		return charIDs, nil
	case model.TargetSelector_ANY_CHARACTER_WITH_ROLE:
		var charIDs []int32
		for id, char := range state.Characters {
			if char.HiddenRole == selector.RoleFilter {
				charIDs = append(charIDs, id)
			}
		}
		return charIDs, nil
	default:
		return nil, fmt.Errorf("unsupported selector type: %s", selector.SelectorType)
	}
}

func (ge *GameEngine) executeEffect(state *model.GameState, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload) ([]*model.GameEvent, error) {
	switch t := effect.EffectType.(type) {
	case *model.Effect_AdjustStat:
		return ge.executeAdjustStatEffect(state, t.AdjustStat, payload)
	case *model.Effect_MoveCharacter:
		return ge.executeMoveCharacterEffect(state, t.MoveCharacter, payload)
	case *model.Effect_AddTrait:
		return ge.executeAddTraitEffect(state, t.AddTrait, payload)
	case *model.Effect_RemoveTrait:
		return ge.executeRemoveTraitEffect(state, t.RemoveTrait, payload)
	case *model.Effect_CompoundEffect:
		return ge.executeCompoundEffect(state, t.CompoundEffect, ability, payload)
	default:
		return nil, fmt.Errorf("unknown effect type: %T", t)
	}
}

func (ge *GameEngine) executeAdjustStatEffect(state *model.GameState, effect *model.AdjustStatEffect, payload *model.UseAbilityPayload) ([]*model.GameEvent, error) {
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, payload)
	if err != nil {
		return nil, err
	}

	var events []*model.GameEvent
	for _, targetID := range targetIDs {
		switch effect.StatType {
		case model.StatCondition_PARANOIA:
			ge.AdjustCharacterParanoia(targetID, effect.Amount)
		case model.StatCondition_INTRIGUE:
			ge.AdjustCharacterIntrigue(targetID, effect.Amount)
		case model.StatCondition_GOODWILL:
			ge.AdjustCharacterGoodwill(targetID, effect.Amount)
		}
	}
	return events, nil
}

func (ge *GameEngine) executeMoveCharacterEffect(state *model.GameState, effect *model.MoveCharacterEffect, payload *model.UseAbilityPayload) ([]*model.GameEvent, error) {
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, payload)
	if err != nil {
		return nil, err
	}
	for _, targetID := range targetIDs {
		ge.SetCharacterLocation(targetID, effect.Destination)
	}
	return nil, nil
}

func (ge *GameEngine) executeAddTraitEffect(state *model.GameState, effect *model.AddTraitEffect, payload *model.UseAbilityPayload) ([]*model.GameEvent, error) {
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, payload)
	if err != nil {
		return nil, err
	}
	for _, targetID := range targetIDs {
		ge.AddCharacterTrait(targetID, effect.Trait)
	}
	return nil, nil
}

func (ge *GameEngine) executeRemoveTraitEffect(state *model.GameState, effect *model.RemoveTraitEffect, payload *model.UseAbilityPayload) ([]*model.GameEvent, error) {
	targetIDs, err := ge.resolveSelectorToCharacters(state, effect.Target, payload)
	if err != nil {
		return nil, err
	}
	for _, targetID := range targetIDs {
		ge.RemoveCharacterTrait(targetID, effect.Trait)
	}
	return nil, nil
}

func (ge *GameEngine) executeCompoundEffect(state *model.GameState, effect *model.CompoundEffect, ability *model.Ability, payload *model.UseAbilityPayload) ([]*model.GameEvent, error) {
	var allEvents []*model.GameEvent
	switch effect.Operator {
	case model.CompoundEffect_SEQUENCE:
		for _, subEffect := range effect.SubEffects {
			events, err := ge.executeEffect(state, subEffect, ability, payload)
			if err != nil {
				return nil, err
			}
			allEvents = append(allEvents, events...)
		}
	case model.CompoundEffect_CHOOSE_ONE:
		choiceID := payload.GetChooseOption().GetChoiceId()
		if !strings.HasPrefix(choiceID, "effect_choice_") {
			return nil, fmt.Errorf("invalid choice id for compound effect: %s", choiceID)
		}
		indexStr := strings.TrimPrefix(choiceID, "effect_choice_")
		choiceIndex, err := strconv.Atoi(indexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid choice index: %s", indexStr)
		}

		if choiceIndex < 0 || choiceIndex >= len(effect.SubEffects) {
			return nil, fmt.Errorf("choice index out of bounds: %d", choiceIndex)
		}
		events, err := ge.executeEffect(state, effect.SubEffects[choiceIndex], ability, payload)
		if err != nil {
			return nil, err
		}
		allEvents = append(allEvents, events...)
	}
	return allEvents, nil
}
