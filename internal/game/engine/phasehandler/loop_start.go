package phasehandler

import (
	"time"
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// --- LoopStartPhase ---
type LoopStartPhase struct{}

// HandleAction is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopStartPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	return nil
}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopStartPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }

// HandleTimeout is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopStartPhase) HandleTimeout(ge GameEngine) Phase { return nil }

// Exit is the default implementation for Phase interface, does nothing.
func (p *LoopStartPhase) Exit(ge GameEngine) {}

// TimeoutDuration is the default implementation for Phase interface, returns 0, indicating no timeout.
func (p *LoopStartPhase) TimeoutDuration() time.Duration { return 0 }

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
