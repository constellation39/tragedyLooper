package effecthandler

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register[*model.Effect_AdjustStat](&AdjustStatHandler{})
}

// AdjustStatHandler 处理 AdjustStat 效果。
type AdjustStatHandler struct{}

func (h *AdjustStatHandler) ResolveChoices(ge GameEngine, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	adjustStatEffect := effect.GetAdjustStat()
	if adjustStatEffect == nil {
		return nil, fmt.Errorf("effect is not of type AdjustStat")
	}
	return CreateChoicesFromSelector(ge, adjustStatEffect.Target, payload, "Select character to adjust stat")
}

func (h *AdjustStatHandler) Apply(ge GameEngine, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	adjustStatEffect := effect.GetAdjustStat()
	if adjustStatEffect == nil {
		return fmt.Errorf("effect is not of type AdjustStat")
	}

	state := ge.GetGameState()
	targetIDs, err := ge.ResolveSelectorToCharacters(state, adjustStatEffect.Target, nil, payload, ability)
	if err != nil {
		return err
	}

	for _, targetID := range targetIDs {
		char, ok := state.Characters[targetID]
		if !ok {
			continue // 或返回错误
		}

		switch adjustStatEffect.StatType {
		case model.StatCondition_PARANOIA:
			newParanoia := char.Paranoia + adjustStatEffect.Amount
			event := &model.ParanoiaAdjustedEvent{CharacterId: targetID, NewParanoia: newParanoia, Amount: adjustStatEffect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_PARANOIA_ADJUSTED, event)
		case model.StatCondition_INTRIGUE:
			newIntrigue := char.Intrigue + adjustStatEffect.Amount
			event := &model.IntrigueAdjustedEvent{CharacterId: targetID, NewIntrigue: newIntrigue, Amount: adjustStatEffect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_INTRIGUE_ADJUSTED, event)
		case model.StatCondition_GOODWILL:
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
	return fmt.Sprintf("调整 %s %d", adjustStat.StatType, adjustStat.Amount)
}