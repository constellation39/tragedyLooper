package eventhandler

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

// registry holds the mapping from event type to its handler.
var registry = make(map[model.GameEventType]EventHandler)

// Register adds an event handler to the registry.
// This function should be called during initialization (e.g., in an init() function).
func Register(eventType model.GameEventType, handler EventHandler) {
	if _, exists := registry[eventType]; exists {
		// Or handle as a fatal error, depending on desired behavior
		panic(fmt.Sprintf("handler for event type %s already registered", eventType))
	}
	registry[eventType] = handler
}

// GetHandler retrieves an event handler from the registry.
func GetHandler(eventType model.GameEventType) (EventHandler, bool) {
	handler, ok := registry[eventType]
	return handler, ok
}
