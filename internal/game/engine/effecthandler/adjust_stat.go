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
	targetIDs, err := ge.ResolveSelectorToCharacters(state, adjustStatEffect.Target, ctx)
	if err != nil {
		return fmt.Errorf("failed to resolve selector to characters: %w", err)
	}

	for _, targetID := range targetIDs {
		if err := h.applyStatAdjustment(ge, targetID, adjustStatEffect); err != nil {
			// Consider whether to continue on error or return immediately
			// For now, we'll log the error and continue
			// logger.Error("failed to apply stat adjustment", "error", err, "targetID", targetID)
			continue
		}
	}
	return nil
}

func (h *AdjustStatHandler) applyStatAdjustment(ge GameEngine, targetID int32, effect *model.AdjustStatEffect) error {
	char, ok := ge.GetGameState().Characters[targetID]
	if !ok {
		return fmt.Errorf("character with id %d not found", targetID)
	}

	var eventType model.GameEventType
	var payload *model.EventPayload

	switch effect.StatType {
	case model.StatCondition_STAT_TYPE_PARANOIA:
		newParanoia := char.Paranoia + effect.Amount
		eventType = model.GameEventType_GAME_EVENT_TYPE_PARANOIA_ADJUSTED
		payload = &model.EventPayload{
			Payload: &model.EventPayload_ParanoiaAdjusted{
				ParanoiaAdjusted: &model.ParanoiaAdjustedEvent{
					CharacterId: targetID,
					NewParanoia: newParanoia,
					Amount:      effect.Amount,
				},
			},
		}

	case model.StatCondition_STAT_TYPE_INTRIGUE:
		newIntrigue := char.Intrigue + effect.Amount
		eventType = model.GameEventType_GAME_EVENT_TYPE_INTRIGUE_ADJUSTED
		payload = &model.EventPayload{
			Payload: &model.EventPayload_IntrigueAdjusted{
				IntrigueAdjusted: &model.IntrigueAdjustedEvent{
					CharacterId: targetID,
					NewIntrigue: newIntrigue,
					Amount:      effect.Amount,
				},
			},
		}

	case model.StatCondition_STAT_TYPE_GOODWILL:
		newGoodwill := char.Goodwill + effect.Amount
		eventType = model.GameEventType_GAME_EVENT_TYPE_GOODWILL_ADJUSTED
		payload = &model.EventPayload{
			Payload: &model.EventPayload_GoodwillAdjusted{
				GoodwillAdjusted: &model.GoodwillAdjustedEvent{
					CharacterId: targetID,
					NewGoodwill: newGoodwill,
					Amount:      effect.Amount,
				},
			},
		}

	default:
		return fmt.Errorf("unknown stat type: %s", effect.StatType)
	}

	ge.ApplyAndPublishEvent(eventType, payload)
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
