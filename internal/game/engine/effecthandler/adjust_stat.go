package effecthandler // Defines the package for effect handlers

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

// init automatically executes when the package is loaded, registering the AdjustStat effect handler.
func init() {
	Register[*model.Effect_AdjustStat](&AdjustStatHandler{})
}

// AdjustStatHandler implements the logic for handling the AdjustStat effect.
// The AdjustStat effect is used to adjust a specified character's stat (e.g., paranoia, intrigue, goodwill).
type AdjustStatHandler struct{}

func (h *AdjustStatHandler) ResolveChoices(ge GameEngine, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	adjustStatEffect := effect.GetAdjustStat()
	if adjustStatEffect == nil {
		return nil, fmt.Errorf("effect is not of type AdjustStat")
	}
	// Create choices from the effect's target selector, allowing the player to choose which character's stat to adjust.
	return CreateChoicesFromSelector(ge, adjustStatEffect.Target, payload, "Select character to adjust stat")
}

func (h *AdjustStatHandler) Apply(ge GameEngine, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	adjustStatEffect := effect.GetAdjustStat()
	if adjustStatEffect == nil {
		return fmt.Errorf("effect is not of type AdjustStat")
	}

	state := ge.GetGameState()
	// Resolve the target selector to get all affected character IDs.
	targetIDs, err := ge.ResolveSelectorToCharacters(state, adjustStatEffect.Target, nil, payload, ability)
	if err != nil {
		return err
	}

	// Iterate over all target characters, adjust the corresponding stat based on StatType, and publish the corresponding event.
	for _, targetID := range targetIDs {
		char, ok := state.Characters[targetID]
		if !ok {
			continue // Or return an error
		}

		switch adjustStatEffect.StatType {
		case model.StatCondition_PARANOIA:
			// Adjust paranoia and publish ParanoiaAdjustedEvent.
			newParanoia := char.Paranoia + adjustStatEffect.Amount
			event := &model.ParanoiaAdjustedEvent{CharacterId: targetID, NewParanoia: newParanoia, Amount: adjustStatEffect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_PARANOIA_ADJUSTED, event)
		case model.StatCondition_INTRIGUE:
			// Adjust intrigue and publish IntrigueAdjustedEvent.
			newIntrigue := char.Intrigue + adjustStatEffect.Amount
			event := &model.IntrigueAdjustedEvent{CharacterId: targetID, NewIntrigue: newIntrigue, Amount: adjustStatEffect.Amount}
			ge.ApplyAndPublishEvent(model.GameEventType_INTRIGUE_ADJUSTED, event)
		case model.StatCondition_GOODWILL:
			// Adjust goodwill and publish GoodwillAdjustedEvent.
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
		return "(Invalid AdjustStat effect)"
	}
	// 返回 AdjustStat 效果的描述字符串。
	return fmt.Sprintf("Adjust %s by %d", adjustStat.StatType, adjustStat.Amount)
}