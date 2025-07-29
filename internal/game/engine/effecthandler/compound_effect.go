package effecthandler // 定义效果处理器的包

import (
	"fmt"
	"strconv"
	"strings"
	model "tragedylooper/pkg/proto/v1"
)

// init 函数在包加载时自动执行，用于注册 CompoundEffect 效果处理器。
func init() {
	Register[*model.Effect_CompoundEffect](&CompoundEffectHandler{})
}

// CompoundEffectHandler 结构体实现了处理复合效果的逻辑。
// 复合效果可以包含多个子效果，并根据操作符（如序列或选择其一）来执行。
type CompoundEffectHandler struct{}

func (h *CompoundEffectHandler) ResolveChoices(ge GameEngine, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	compoundEffect := effect.GetCompoundEffect()
	if compoundEffect == nil {
		return nil, fmt.Errorf("effect is not of type CompoundEffect")
	}

	switch compoundEffect.Operator {
	case model.CompoundEffect_CHOOSE_ONE:
		// 如果是 CHOOSE_ONE 类型，为每个子效果创建一个选择项。
		var choices []*model.Choice
		for i, subEffect := range compoundEffect.SubEffects {
			choiceID := fmt.Sprintf("effect_choice_%d", i)
			choices = append(choices, &model.Choice{
				Id:          choiceID,
				Description: GetEffectDescription(ge, subEffect), // 我们需要一种方法来获取描述
				ChoiceType:  &model.Choice_EffectOptionIndex{EffectOptionIndex: int32(i)},
			})
		}
		return choices, nil
	case model.CompoundEffect_SEQUENCE:
		// 如果是 SEQUENCE 类型，按顺序解析子效果的选择项，直到找到第一个需要选择的效果。
		for _, subEffect := range compoundEffect.SubEffects {
			// 在序列中，我们呈现的第一个需要选择的效果。
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
		// 如果是 SEQUENCE 类型，按顺序应用所有子效果。
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
		// 如果是 CHOOSE_ONE 类型，根据玩家的选择应用对应的子效果。
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
	// 返回复合效果的描述字符串。
	return "Choose one of the following effects"
}