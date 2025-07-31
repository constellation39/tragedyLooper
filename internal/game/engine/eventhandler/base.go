package eventhandler

import (
	"fmt"
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// PhaseManager defines the interface for a component that manages game phases
// and can react to events. This is implemented by engine.PhaseManager.
type PhaseManager interface {
	HandleEvent(event *model.GameEvent)
}

// GameEngine defines the dependencies that event handlers and the event manager need
// from the main game engine. It's an interface to decouple the packages.
type GameEngine interface {
	GetGameState() *model.GameState
	Logger() *zap.Logger
	GetPhaseManager() PhaseManager
}

// EventHandler defines the interface for handling a game event.
// It modifies the game state directly based on the event.
type EventHandler interface {
	Handle(ge GameEngine, event *model.GameEvent) error
}

// registry is a map of event types to their corresponding handlers.
var registry = make(map[model.GameEventType]EventHandler)

// Register associates an event handler with a game event type.
// It is called from the init() function of each event handler implementation.
func Register(eventType model.GameEventType, handler EventHandler) {
	if _, exists := registry[eventType]; exists {
		// This should not happen in production, so we panic.
		panic(fmt.Sprintf("handler for event type %s already registered", eventType))
	}
	registry[eventType] = handler
}

// GetHandler retrieves the handler for a given event type.
// It returns the handler and a boolean indicating if a handler was found.
func GetHandler(eventType model.GameEventType) (EventHandler, bool) {
	handler, ok := registry[eventType]
	return handler, ok
}