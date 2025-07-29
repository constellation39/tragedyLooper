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

// ProcessEvent finds the appropriate handler in the registry and uses it to process the event.
// It returns an error if no handler is found or if the handler itself returns an error.
func ProcessEvent(state *model.GameState, event *model.EventPayload) error {
	handler, ok := registry[event.Type]
	if !ok {
		return fmt.Errorf("no handler registered for event type %s", event.Type)
	}
	return handler.Handle(state, event)
}
