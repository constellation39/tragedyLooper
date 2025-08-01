package phasehandler

import (
	"time"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// --- LoopStartPhase ---
type LoopStartPhase struct{}

// HandleAction is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopStartPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) {
}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopStartPhase) HandleEvent(ge GameEngine, event *model.GameEvent) {}

// HandleTimeout is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopStartPhase) HandleTimeout(ge GameEngine) {}

// Exit is the default implementation for Phase interface, does nothing.
func (p *LoopStartPhase) Exit(ge GameEngine) {}

// TimeoutDuration is the default implementation for Phase interface, returns 0, indicating no timeout.
func (p *LoopStartPhase) TimeoutDuration() time.Duration { return 0 }

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
