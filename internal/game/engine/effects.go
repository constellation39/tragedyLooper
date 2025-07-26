package engine

import (
	"fmt"
	"tragedylooper/internal/game/proto/model"
)

func (ge *GameEngine) applyEffect(effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload) error {
	ctx := &model.EffectContext{GameState: ge.GameState}

	// First, see if the effect requires a choice from the player.
	choices, err := resolveEffectChoices(ctx, effect, ability)
	if err != nil {
		return fmt.Errorf("error resolving choices: %w", err)
	}

	// If choices are available and no specific target was provided in the payload, ask the player.
	// A more robust check might be needed, e.g., checking if payload.Target is fully specified.
	if len(choices) > 1 && (payload.Target == nil || payload.Target.GetCharacterId() == 0) {
		choiceEvent := &model.ChoiceRequiredEvent{Choices: choices}
		ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_CHOICE_REQUIRED, choiceEvent)
		return nil // Stop processing and wait for a player action with the choice.
	}

	// If a choice was made or not required, execute the effect.
	events, err := executeEffect(ge, ctx, effect, ability, payload)
	if err != nil {
		return fmt.Errorf("error executing effect: %w", err)
	}

	// Apply all resulting events to the game state.
	for _, event := range events {
		ge.processEvent(event)
	}

	return nil
}

func resolveEffectChoices(ctx *model.EffectContext, effect *model.Effect, ability *model.Ability) ([]*model.Choice, error) {
	switch t := effect.EffectOneof.(type) {
	case *model.Effect_MoveCharacterEffect:
		// No choices needed for this effect
		return nil, nil
	case *model.Effect_AdjustParanoiaEffect:
		// No choices needed for this effect
		return nil, nil
	case *model.Effect_AdjustGoodwillEffect:
		// No choices needed for this effect
		return nil, nil
	case *model.Effect_AdjustIntrigueEffect:
		// No choices needed for this effect
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown effect type: %T", t)
	}
}

func executeEffect(ge *GameEngine, ctx *model.EffectContext, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload) ([]*model.GameEvent, error) {
	switch t := effect.EffectOneof.(type) {
	case *model.Effect_MoveCharacterEffect:
		ge.SetCharacterLocation(payload.GetCharacterId(), t.MoveCharacterEffect.Destination)
		return nil, nil
	case *model.Effect_AdjustParanoiaEffect:
		ge.AdjustCharacterParanoia(payload.GetCharacterId(), t.AdjustParanoiaEffect.Amount)
		return nil, nil
	case *model.Effect_AdjustGoodwillEffect:
		ge.AdjustCharacterGoodwill(payload.GetCharacterId(), t.AdjustGoodwillEffect.Amount)
		return nil, nil
	case *model.Effect_AdjustIntrigueEffect:
		ge.AdjustCharacterIntrigue(payload.GetCharacterId(), t.AdjustIntrigueEffect.Amount)
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown effect type: %T", t)
	}
}
