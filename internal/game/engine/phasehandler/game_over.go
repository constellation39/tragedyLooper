package phasehandler

import (
	"time"
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// --- GameOverPhase ---
type GameOverPhase struct{}

// HandleAction is the default implementation for Phase interface, does nothing and returns nil.
func (p *GameOverPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	return nil
}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *GameOverPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }

// HandleTimeout is the default implementation for Phase interface, does nothing and returns nil.
func (p *GameOverPhase) HandleTimeout(ge GameEngine) Phase { return nil }

// Exit is the default implementation for Phase interface, does nothing.
func (p *GameOverPhase) Exit(ge GameEngine) {}

// TimeoutDuration is the default implementation for Phase interface, returns 0, indicating no timeout.
func (p *GameOverPhase) TimeoutDuration() time.Duration { return 0 }


func (p *GameOverPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_GAME_OVER }
func (p *GameOverPhase) Enter(ge GameEngine) Phase {
	// Clean up, announce winner, etc.
	return nil
}

func init() {
	RegisterPhase(&GameOverPhase{})
}
