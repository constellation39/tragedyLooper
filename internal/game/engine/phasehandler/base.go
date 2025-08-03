package phasehandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// BasePhase provides a default implementation for the Phase interface.
// It is intended to be embedded in concrete phase implementations.
// This way, each phase only needs to implement the methods relevant to it,
// reducing boilerplate code.
type BasePhase struct {
	readyToTransition bool
}

// Enter is a default implementation that does nothing.
func (p *BasePhase) Enter(ge GameEngine) PhaseState {
	return PhaseComplete
}

// HandleAction is a default implementation that does nothing and indicates that the phase is not ready to transition.
func (p *BasePhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) PhaseState {
	return PhaseComplete
}

// HandleEvent is a default implementation that does nothing and indicates that the phase is not ready to transition.
func (p *BasePhase) HandleEvent(ge GameEngine, event *model.GameEvent) PhaseState {
	return PhaseComplete
}

// HandleTimeout is a default implementation that does nothing.
func (p *BasePhase) HandleTimeout(ge GameEngine) {}

// Exit is a default implementation that does nothing.
func (p *BasePhase) Exit(ge GameEngine) {}

// TimeoutTicks is a default implementation that returns 0, indicating no timeout.
func (p *BasePhase) TimeoutTicks() int64 { return 0 }
