package eventhandler

import model "tragedylooper/pkg/proto/v1"

// EventHandler defines the interface for handling a game event.
// It modifies the game state directly based on the event.
type EventHandler interface {
	Handle(state *model.GameState, event *model.EventPayload) error
}
