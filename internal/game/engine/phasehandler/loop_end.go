package phasehandler

import (
	
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// --- LoopEndPhase ---
type LoopEndPhase struct{
	BasePhase
}



func (p *LoopEndPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_LOOP_END }
func (p *LoopEndPhase) Enter(ge GameEngine) {
	gs := ge.GetGameState()
	script := ge.GetGameRepo().GetScript()

	if gs.CurrentLoop >= script.LoopCount {
		// Final loop has ended. Check for protagonist win condition.
		// This is a simplification. A real game would have more complex win/loss checks.
		ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_LOOP_WIN, &model.EventPayload{
			Payload: &model.EventPayload_LoopWin{LoopWin: &model.LoopWinEvent{}},
		})
	} else {
		// Reset for the next loop
		ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_LOOP_RESET, &model.EventPayload{
			Payload: &model.EventPayload_LoopReset{LoopReset: &model.LoopResetEvent{LoopNumber: gs.CurrentLoop + 1}},
		})
	}
}

func init() {
	RegisterPhase(&LoopEndPhase{})
}
