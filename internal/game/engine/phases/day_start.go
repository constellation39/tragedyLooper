package phases

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- DayStartPhase ---
type DayStartPhase struct{ basePhase }

func (p *DayStartPhase) Type() model.GamePhase { return model.GamePhase_DAY_START }
func (p *DayStartPhase) Enter(ge GameEngine) Phase {
	ge.GetGameState().CurrentDay++
	ge.GetGameState().PlayedCardsThisDay = make(map[int32]*model.Card)
	ge.ResetPlayerReadiness()
	ge.ApplyAndPublishEvent(model.GameEventType_DAY_ADVANCED, &model.DayAdvancedEvent{Day: ge.GetGameState().CurrentDay, Loop: ge.GetGameState().CurrentLoop})
	return &CardPlayPhase{}
}
