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
	case model.CompoundEffect_OPERATOR_CHOOSE_ONE:
		return h.resolveChooseOneChoices(ge, compoundEffect)
	case model.CompoundEffect_OPERATOR_SEQUENCE:
		return h.resolveSequenceChoices(ge, compoundEffect, ctx)
	default:
		return nil, fmt.Errorf("unknown compound effect operator: %v", compoundEffect.Operator)
	}
}

func (h *CompoundEffectHandler) resolveChooseOneChoices(ge GameEngine, compoundEffect *model.CompoundEffect) ([]*model.Choice, error) {
	if len(compoundEffect.SubEffects) >= math.MaxInt32 {
		return nil, fmt.Errorf("too many sub-effects, exceeds int32 range")
	}
	var choices []*model.Choice
	for i, subEffect := range compoundEffect.SubEffects {
		choiceID := fmt.Sprintf("effect_choice_%d", i)
		choices = append(choices, &model.Choice{
			Id:          choiceID,
			Description: GetEffectDescription(ge, subEffect),
			ChoiceType:  &model.Choice_EffectOptionIndex{EffectOptionIndex: int32(i)}, //nolint:gosec
		})
	}
	return choices, nil
}

func (h *CompoundEffectHandler) resolveSequenceChoices(ge GameEngine, compoundEffect *model.CompoundEffect, ctx *EffectContext) ([]*model.Choice, error) {
	for _, subEffect := range compoundEffect.SubEffects {
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
	return nil, nil
}

func (h *CompoundEffectHandler) Apply(ge GameEngine, effect *model.Effect, ctx *EffectContext) error {
	compoundEffect := effect.GetCompoundEffect()
	if compoundEffect == nil {
		return fmt.Errorf("effect is not of type CompoundEffect")
	}

	switch compoundEffect.Operator {
	case model.CompoundEffect_OPERATOR_SEQUENCE:
		return h.applySequence(ge, compoundEffect, ctx)
	case model.CompoundEffect_OPERATOR_CHOOSE_ONE:
		return h.applyChooseOne(ge, compoundEffect, ctx)
	default:
		return fmt.Errorf("unknown compound effect operator: %v", compoundEffect.Operator)
	}
}

func (h *CompoundEffectHandler) applySequence(ge GameEngine, compoundEffect *model.CompoundEffect, ctx *EffectContext) error {
	for _, subEffect := range compoundEffect.SubEffects {
		if err := ApplyEffect(ge, subEffect, ctx); err != nil {
			return fmt.Errorf("failed to apply sub-effect in sequence: %w", err)
		}
	}
	return nil
}

func (h *CompoundEffectHandler) applyChooseOne(ge GameEngine, compoundEffect *model.CompoundEffect, ctx *EffectContext) error {
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
	return ApplyEffect(ge, chosenEffect, ctx)
}

func (h *CompoundEffectHandler) GetDescription(effect *model.Effect) string {
	// 返回复合效果的描述字符串。
	return "Choose one of the following effects"
}
