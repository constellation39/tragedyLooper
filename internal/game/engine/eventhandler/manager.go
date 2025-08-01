package eventhandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// Manager is responsible for creating, processing, and dispatching all game events.
// It decouples the event lifecycle from the main GameEngine, ensuring a clear and maintainable flow.
type Manager struct {
	engine GameEngine // Reference to the parent engine to access global state and managers.
	logger *zap.Logger

	// dispatchGameEvent is an outbound channel for broadcasting processed game events to external listeners.
	dispatchGameEvent chan *model.GameEvent
}

// NewManager creates and returns a new eventManager instance.
// engine: A reference to the game engine.
func NewManager(engine GameEngine) *Manager {
	return &Manager{
		engine:            engine,
		logger:            engine.Logger().Named("EventManager"),
		dispatchGameEvent: make(chan *model.GameEvent, 100),
	}
}

func (em *Manager) ApplyEvent(event *model.GameEvent) error {
	handler, ok := GetHandler(event.Type)
	if !ok {
		// Not all events have handlers, so this is not an error.
		// It simply means no state change is associated with this event by default.
		return nil
	}

	return handler.Handle(em.engine, event)
}

func (em *Manager) Dispatch(event *model.GameEvent) {
	select {
	case em.dispatchGameEvent <- event:
		// Event successfully dispatched.
	default:
		em.logger.Warn("Game event channel full, dropping event", zap.String("eventType", event.Type.String()))
	}
}

// EventsChannel returns the outbound channel for game events.
// Returns: A read-only channel of game events.
func (em *Manager) EventsChannel() <-chan *model.GameEvent {
	return em.dispatchGameEvent
}

// Close closes the event dispatch channel.
func (em *Manager) Close() {
	close(em.dispatchGameEvent)
}
