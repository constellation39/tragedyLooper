package effecthandler // 效果处理器

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// init 函数在包加载时自动执行，注册复合效果处理器。
func init() {
	Register[*model.Effect_CompoundEffect](&CompoundEffectHandler{})
}

// CompoundEffectHandler 实现处理复合效果的逻辑。
// 复合效果可以包含多个子效果，并根据操作符（例如，序列或选择其一）来执行它们。
type CompoundEffectHandler struct{}

func (h *CompoundEffectHandler) ResolveChoices(ge GameEngine, effect *model.Effect, ctx *EffectContext) ([]*model.Choice, error) {
	compoundEffect := effect.GetCompoundEffect()
	if compoundEffect == nil {
		return nil, fmt.Errorf("effect is not of type CompoundEffect")
	}

	switch compoundEffect.Operator {
	case model.CompoundEffect_CHOOSE_ONE:
		// 如果是 CHOOSE_ONE 类型，为每个子效果创建一个选项。
		if len(compoundEffect.SubEffects) >= math.MaxInt32 {
			return nil, fmt.Errorf("too many sub-effects, exceeds int32 range")
		}
		var choices []*model.Choice
		for i, subEffect := range compoundEffect.SubEffects {
			choiceID := fmt.Sprintf("effect_choice_%d", i)
			choices = append(choices, &model.Choice{
				Id:          choiceID,
				Description: GetEffectDescription(ge, subEffect),                          // 我们需要一种方法来获取描述
				ChoiceType:  &model.Choice_EffectOptionIndex{EffectOptionIndex: int32(i)}, //nolint:gosec
			})
		}
		return choices, nil
	case model.CompoundEffect_SEQUENCE:
		// 如果是 SEQUENCE 类型，按顺序为子效果解析选项，直到找到第一个需要选择的子效果。
		for _, subEffect := range compoundEffect.SubEffects {
			// 在序列中，我们提供第一个需要选择的效果。
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
		// 如果是 SEQUENCE 类型，按顺序应用所有子效果。
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
		// 如果是 CHOOSE_ONE 类型，根据玩家的选择应用相应的子效果。
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
	// 返回复合效果的描述字符串。
	return "Choose one of the following effects"
}
