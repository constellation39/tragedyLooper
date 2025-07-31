package phase

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// --- LoopStartPhase ---
type LoopStartPhase struct{ basePhase }

func (p *LoopStartPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_LOOP_START }
func (p *LoopStartPhase) Enter(ge GameEngine) Phase {
	ge.GetGameState().CurrentLoop++
	ge.GetGameState().CurrentDay = 0
	ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_LOOP_RESET, &model.EventPayload{
		Payload: &model.EventPayload_LoopReset{LoopReset: &model.LoopResetEvent{LoopNumber: ge.GetGameState().CurrentLoop}},
	})
	return GetPhase(model.GamePhase_GAME_PHASE_DAY_START)
}

func init() {
	RegisterPhase(&LoopStartPhase{})
}
