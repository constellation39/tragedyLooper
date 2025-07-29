package effecthandler // 定义效果处理器的包

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

// init 函数在包加载时自动执行，用于注册 AdjustStat 效果处理器。
func init() {
	Register[*model.Effect_AdjustStat](&AdjustStatHandler{})
}

// AdjustStatHandler 结构体实现了处理 AdjustStat 效果的逻辑。
// AdjustStat 效果用于调整指定角色的某个统计值（如偏执、阴谋、好感）。
type AdjustStatHandler struct{}

func (h *AdjustStatHandler) ResolveChoices(ge GameEngine, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	adjustStatEffect := effect.GetAdjustStat()
	if adjustStatEffect == nil {
		return nil, fmt.Errorf("effect is not of type AdjustStat")
	}
	// 根据效果的目标选择器创建选项，让玩家选择要调整统计值的角色。
	return CreateChoicesFromSelector(ge, adjustStatEffect.Target, payload, "Select character to adjust stat")
}

func (h *AdjustStatHandler) Apply(ge GameEngine, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	adjustStatEffect := effect.GetAdjustStat()
	if adjustStatEffect == nil {
		return fmt.Errorf("effect is not of type AdjustStat")
	}

	state := ge.GetGameState()
	// 解析目标选择器，获取所有受影响的角色ID。
	targetIDs, err := ge.ResolveSelectorToCharacters(state, adjustStatEffect.Target, nil, payload, ability)
	if err != nil {
		return err
	}

	// 遍历所有目标角色，根据 StatType 调整对应的统计值，并发布相应的事件。
	for _, targetID := range targetIDs {
		char, ok := state.Characters[targetID]
		if !ok {
			continue // 或返回错误
		}

		switch adjustStatEffect.StatType {
		case model.StatCondition_PARANOIA:
			// 调整偏执值并发布 ParanoiaAdjustedEvent 事件。
			newParanoia := char.Paranoia + adjustStatEffect.Amount
			event := &model.ParanoiaAdjustedEvent{CharacterId: targetID, NewParanoia: newParanoia, Amount: adjustStatEffect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_PARANOIA_ADJUSTED, event)
		case model.StatCondition_INTRIGUE:
			// 调整阴谋值并发布 IntrigueAdjustedEvent 事件。
			newIntrigue := char.Intrigue + adjustStatEffect.Amount
			event := &model.IntrigueAdjustedEvent{CharacterId: targetID, NewIntrigue: newIntrigue, Amount: adjustStatEffect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_INTRIGUE_ADJUSTED, event)
		case model.StatCondition_GOODWILL:
			// 调整好感值并发布 GoodwillAdjustedEvent 事件。
			newGoodwill := char.Goodwill + adjustStatEffect.Amount
			event := &model.GoodwillAdjustedEvent{CharacterId: targetID, NewGoodwill: newGoodwill, Amount: adjustStatEffect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_GOODWILL_ADJUSTED, event)
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
	return fmt.Sprintf("调整 %s %d", adjustStat.StatType, adjustStat.Amount)
}