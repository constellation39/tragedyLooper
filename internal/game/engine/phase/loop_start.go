package phase

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- LoopStartPhase ---
type LoopStartPhase struct{ basePhase }

func (p *LoopStartPhase) Type() model.GamePhase { return model.GamePhase_LOOP_START }
func (p *LoopStartPhase) Enter(ge GameEngine) Phase {
	ge.GetGameState().CurrentLoop++
	ge.GetGameState().CurrentDay = 0
	ge.ApplyAndPublishEvent(model.GameEventType_LOOP_RESET, &model.EventPayload{
		Payload: &model.EventPayload_LoopReset{LoopReset: &model.LoopResetEvent{LoopNumber: ge.GetGameState().CurrentLoop}},
	})
	return &DayStartPhase{}
}
