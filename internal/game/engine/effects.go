package engine

import (
	"fmt"
	"tragedylooper/internal/game/engine/effecthandler"
	model "tragedylooper/pkg/proto/v1"
)

// applyEffect finds the appropriate handler for an effect and uses it to resolve choices and then apply the effect.
func (ge *GameEngine) applyEffect(effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	handler, err := effecthandler.GetEffectHandler(effect)
	if err != nil {
		return err
	}

	// 1. Resolve choices
	// If choices are required, publish an event and wait for player input.
	// The actual application of the effect will happen once the choice is made.
	choices, err := handler.ResolveChoices(ge, effect, payload)
	if err != nil {
		return fmt.Errorf("error resolving choices: %w", err)
	}

	if len(choices) > 0 && choice == nil {
		choiceEvent := &model.ChoiceRequiredEvent{Choices: choices}
		ge.ApplyAndPublishEvent(model.GameEventType_CHOICE_REQUIRED, choiceEvent)
		return nil // Stop processing until a choice is received
	}

	// 2. Apply the effect
	// If no choices are needed, or if a choice has been provided, apply the effect.
	err = handler.Apply(ge, effect, ability, payload, choice)
	if err != nil {
		return fmt.Errorf("error applying effect: %w", err)
	}

	return nil
}

// GetEffectDescription finds the appropriate handler and returns the effect's description.
func (ge *GameEngine) GetEffectDescription(effect *model.Effect) string {
	return effecthandler.GetEffectDescription(ge, effect)
}
