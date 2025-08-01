package phasehandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// --- LoopStartPhase ---
type LoopStartPhase struct {
	BasePhase
}

func (p *LoopStartPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_LOOP_START }
func (p *LoopStartPhase) Enter(ge GameEngine) {
	gs := ge.GetGameState()
	script := ge.GetGameRepo().GetScript()

	if gs.CurrentLoop >= script.LoopCount {
		ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_GAME_ENDED, &model.EventPayload{
			Payload: &model.EventPayload_GameEnded{GameEnded: &model.GameEndedEvent{Reason: "Max loops reached"}},
		})
		return
	}

	gs.CurrentLoop++
	gs.CurrentDay = 0
	ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_LOOP_RESET, &model.EventPayload{
		Payload: &model.EventPayload_LoopReset{LoopReset: &model.LoopResetEvent{LoopNumber: gs.CurrentLoop}},
	})
}

func init() {
	RegisterPhase(&LoopStartPhase{})
}
