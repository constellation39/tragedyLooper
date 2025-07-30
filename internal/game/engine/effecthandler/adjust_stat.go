package effecthandler

import (
	"fmt"
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// init 函数在包加载时自动执行，注册 AdjustStat 效果处理器。
func init() {
	Register[*model.Effect_AdjustStat](&AdjustStatHandler{})
}

// AdjustStatHandler 实现处理 AdjustStat 效果的逻辑。
// AdjustStat 效果用于调整指定角色的属性（例如，偏执、阴谋、好感）。
type AdjustStatHandler struct{}

func (h *AdjustStatHandler) ResolveChoices(ge GameEngine, effect *model.Effect, ctx *EffectContext) ([]*model.Choice, error) {
	adjustStatEffect := effect.GetAdjustStat()
	if adjustStatEffect == nil {
		return nil, fmt.Errorf("effect is not of type AdjustStat")
	}
	// 根据效果的目标选择器创建选项，让玩家选择要调整哪个角色的属性。
	return CreateChoicesFromSelector(ge, adjustStatEffect.Target, ctx, "Select character to adjust stat")
}

func (h *AdjustStatHandler) Apply(ge GameEngine, effect *model.Effect, ctx *EffectContext) error {
	adjustStatEffect := effect.GetAdjustStat()
	if adjustStatEffect == nil {
		return fmt.Errorf("effect is not of type AdjustStat")
	}

	state := ge.GetGameState()
	// 解析目标选择器以获取所有受影响的角色 ID。
	targetIDs, err := ge.ResolveSelectorToCharacters(state, adjustStatEffect.Target, ctx)
	if err != nil {
		return err
	}

	// 遍历所有目标角色，根据 StatType 调整相应的属性，并发布相应的事件。
	for _, targetID := range targetIDs {
		char, ok := state.Characters[targetID]
		if !ok {
			continue // 或返回错误
		}

		switch adjustStatEffect.StatType {
		case model.StatCondition_STAT_TYPE_PARANOIA:
			// 调整偏执并发布 ParanoiaAdjustedEvent。
			newParanoia := char.Paranoia + adjustStatEffect.Amount
			event := &model.ParanoiaAdjustedEvent{CharacterId: targetID, NewParanoia: newParanoia, Amount: adjustStatEffect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_GAME_EVENT_TYPE_PARANOIA_ADJUSTED, &model.EventPayload{
				Payload: &model.EventPayload_ParanoiaAdjusted{ParanoiaAdjusted: event},
			})
		case model.StatCondition_STAT_TYPE_INTRIGUE:
			// 调整阴谋并发布 IntrigueAdjustedEvent。
			newIntrigue := char.Intrigue + adjustStatEffect.Amount
			event := &model.IntrigueAdjustedEvent{CharacterId: targetID, NewIntrigue: newIntrigue, Amount: adjustStatEffect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_GAME_EVENT_TYPE_INTRIGUE_ADJUSTED, &model.EventPayload{
				Payload: &model.EventPayload_IntrigueAdjusted{IntrigueAdjusted: event},
			})
		case model.StatCondition_STAT_TYPE_GOODWILL:
			// 调整好感并发布 GoodwillAdjustedEvent。
			newGoodwill := char.Goodwill + adjustStatEffect.Amount
			event := &model.GoodwillAdjustedEvent{CharacterId: targetID, NewGoodwill: newGoodwill, Amount: adjustStatEffect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_GAME_EVENT_TYPE_GOODWILL_ADJUSTED, &model.EventPayload{
				Payload: &model.EventPayload_GoodwillAdjusted{GoodwillAdjusted: event},
			})
		}
	}
	return nil
}

func (h *AdjustStatHandler) GetDescription(effect *model.Effect) string {
	adjustStat := effect.GetAdjustStat()
	if adjustStat == nil {
		return "(无效的 AdjustStat 效果)"
	}
	// 返回 AdjustStat 效果的描述字符串。
	return fmt.Sprintf("Adjust %s by %d", adjustStat.StatType, adjustStat.Amount)
}
