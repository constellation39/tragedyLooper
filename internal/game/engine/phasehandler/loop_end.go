package phasehandler

import (
	"time"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// --- LoopEndPhase ---
type LoopEndPhase struct{}

// HandleAction is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopEndPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	return nil
}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopEndPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }

// HandleTimeout is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopEndPhase) HandleTimeout(ge GameEngine) Phase { return nil }

// Exit is the default implementation for Phase interface, does nothing.
func (p *LoopEndPhase) Exit(ge GameEngine) {}

// TimeoutDuration is the default implementation for Phase interface, returns 0, indicating no timeout.
func (p *LoopEndPhase) TimeoutDuration() time.Duration { return 0 }

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
