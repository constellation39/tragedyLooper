package effecthandler // Defines the package for effect handlers

import (
	"fmt"
	"strconv"
	"strings"
	model "tragedylooper/pkg/proto/v1"
)

// init automatically executes when the package is loaded, registering the CompoundEffect effect handler.
func init() {
	Register[*model.Effect_CompoundEffect](&CompoundEffectHandler{})
}

// CompoundEffectHandler implements the logic for handling compound effects.
// A compound effect can contain multiple sub-effects and executes them based on an operator (e.g., sequence or choose one).
type CompoundEffectHandler struct{}

func (h *CompoundEffectHandler) ResolveChoices(ge GameEngine, effect *model.Effect, ctx *EffectContext) ([]*model.Choice, error) {
	compoundEffect := effect.GetCompoundEffect()
	if compoundEffect == nil {
		return nil, fmt.Errorf("effect is not of type CompoundEffect")
	}

	switch compoundEffect.Operator {
	case model.CompoundEffect_CHOOSE_ONE:
		// If it's a CHOOSE_ONE type, create a choice for each sub-effect.
		var choices []*model.Choice
		for i, subEffect := range compoundEffect.SubEffects {
			choiceID := fmt.Sprintf("effect_choice_%d", i)
			choices = append(choices, &model.Choice{
				Id:          choiceID,
				Description: GetEffectDescription(ge, subEffect), // We need a way to get the description
				ChoiceType:  &model.Choice_EffectOptionIndex{EffectOptionIndex: int32(i)},
			})
		}
		return choices, nil
	case model.CompoundEffect_SEQUENCE:
		// If it's a SEQUENCE type, resolve choices for sub-effects in order until the first one that requires a choice is found.
		for _, subEffect := range compoundEffect.SubEffects {
			// In a sequence, we present the first effect that requires a choice.
			handler, err := GetEffectHandler(subEffect)
			if err != nil {
				return nil, err
			}
			choices, err := handler.ResolveChoices(ge, subEffect, ctx)
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

func (h *CompoundEffectHandler) Apply(ge GameEngine, effect *model.Effect, ctx *EffectContext) error {
	compoundEffect := effect.GetCompoundEffect()
	if compoundEffect == nil {
		return fmt.Errorf("effect is not of type CompoundEffect")
	}

	switch compoundEffect.Operator {
	case model.CompoundEffect_SEQUENCE:
		// If it's a SEQUENCE type, apply all sub-effects in order.
		for _, subEffect := range compoundEffect.SubEffects {
			handler, err := GetEffectHandler(subEffect)
			if err != nil {
				return err
			}
			err = handler.Apply(ge, subEffect, ctx)
			if err != nil {
				return err
			}
		}
	case model.CompoundEffect_CHOOSE_ONE:
		// If it's a CHOOSE_ONE type, apply the corresponding sub-effect based on the player's choice.
		if ctx == nil || ctx.Choice == nil {
			return fmt.Errorf("a choice is required to apply a CHOOSE_ONE compound effect")
		}
		choiceID := ctx.Choice.GetChosenOptionId()
		if !strings.HasPrefix(choiceID, "effect_choice_") {
			return fmt.Errorf("invalid choice id for compound effect: %s", choiceID)
		}
		indexStr := strings.TrimPrefix(choiceID, "effect_choice_")
		choiceIndex, err := strconv.Atoi(indexStr)
		if err != nil {
			return fmt.Errorf("invalid choice index: %s", indexStr)
		}

		if choiceIndex < 0 || choiceIndex >= len(compoundEffect.SubEffects) {
			return fmt.Errorf("choice index out of bounds: %d", choiceIndex)
		}

		chosenEffect := compoundEffect.SubEffects[choiceIndex]
		handler, err := GetEffectHandler(chosenEffect)
		if err != nil {
			return err
		}
		return handler.Apply(ge, chosenEffect, ctx)
	}
	return nil
}

func (h *CompoundEffectHandler) GetDescription(effect *model.Effect) string {
	// Returns the description string for the compound effect.
	return "Choose one of the following effects"
}
