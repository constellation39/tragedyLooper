package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// EventHandler defines the interface for handling a game event.
// It modifies the game state directly based on the event.
type GameEngine interface {
	GetGameState() *model.GameState
	Logger() *zap.Logger
}

// EventHandler defines the interface for handling a game event.
// It modifies the game state directly based on the event.
type EventHandler interface {
	Handle(ge GameEngine, event *model.GameEvent) error
}
