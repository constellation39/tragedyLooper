package effecthandler

import (
	"fmt"
	"strconv"
	"strings"
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register[*model.Effect_CompoundEffect](&CompoundEffectHandler{})
}

// CompoundEffectHandler processes Compound effects.
type CompoundEffectHandler struct{}

func (h *CompoundEffectHandler) ResolveChoices(ge GameEngine, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	compoundEffect := effect.GetCompoundEffect()
	if compoundEffect == nil {
		return nil, fmt.Errorf("effect is not of type CompoundEffect")
	}

	switch compoundEffect.Operator {
	case model.CompoundEffect_CHOOSE_ONE:
		var choices []*model.Choice
		for i, subEffect := range compoundEffect.SubEffects {
			choiceID := fmt.Sprintf("effect_choice_%d", i)
			choices = append(choices, &model.Choice{
				Id:          choiceID,
				Description: GetEffectDescription(ge, subEffect), // We need a way to get descriptions
				ChoiceType:  &model.Choice_EffectOptionIndex{EffectOptionIndex: int32(i)},
			})
		}
		return choices, nil
	case model.CompoundEffect_SEQUENCE:
		for _, subEffect := range compoundEffect.SubEffects {
			// In a sequence, the first effect that requires a choice is the one we present.
			handler, err := GetEffectHandler(subEffect)
			if err != nil {
				return nil, err
			}
			choices, err := handler.ResolveChoices(ge, subEffect, payload)
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

func (h *CompoundEffectHandler) Apply(ge GameEngine, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	compoundEffect := effect.GetCompoundEffect()
	if compoundEffect == nil {
		return fmt.Errorf("effect is not of type CompoundEffect")
	}

	switch compoundEffect.Operator {
	case model.CompoundEffect_SEQUENCE:
		for _, subEffect := range compoundEffect.SubEffects {
			handler, err := GetEffectHandler(subEffect)
			if err != nil {
				return err
			}
			err = handler.Apply(ge, subEffect, ability, payload, choice)
			if err != nil {
				return err
			}
		}
	case model.CompoundEffect_CHOOSE_ONE:
		if choice == nil {
			return fmt.Errorf("a choice is required to apply a CHOOSE_ONE compound effect")
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

		if choiceIndex < 0 || choiceIndex >= len(compoundEffect.SubEffects) {
			return fmt.Errorf("choice index out of bounds: %d", choiceIndex)
		}

		chosenEffect := compoundEffect.SubEffects[choiceIndex]
		handler, err := GetEffectHandler(chosenEffect)
		if err != nil {
			return err
			}
		return handler.Apply(ge, chosenEffect, ability, payload, choice)
	}
	return nil
}

func (h *CompoundEffectHandler) GetDescription(effect *model.Effect) string {
	return "Choose one of the following effects"
}
