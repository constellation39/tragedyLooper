package engine

import (
	"tragedylooper/internal/game/engine/eventhandler"
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// createAndProcess is the central method for creating, applying, and broadcasting a game event.
// It ensures a consistent order of operations:
// 1. An event is created from a payload.
// 2. The phaseManager is notified, allowing the current phase to react and potentially trigger a transition.
// 3. The event is broadcast to external listeners and recorded in the game's history.
// eventType: The type of the game event.
// payload: The payload for the event, which must be a protobuf message.

// eventManager is responsible for creating, processing, and dispatching all game events.
// It decouples the event lifecycle from the main GameEngine, ensuring a clear and maintainable flow.
type eventManager struct {
	engine *GameEngine // Reference to the parent engine to access global state and managers.
	logger *zap.Logger

	// dispatchGameEvent is an outbound channel for broadcasting processed game events to external listeners.
	dispatchGameEvent chan *model.GameEvent
}

// newEventManager creates and returns a new eventManager instance.
// engine: A reference to the game engine.
func newEventManager(engine *GameEngine) *eventManager {
	return &eventManager{
		engine:            engine,
		logger:            engine.logger.Named("EventManager"),
		dispatchGameEvent: make(chan *model.GameEvent, 100),
	}
}

// createAndProcess is the central method for creating, applying, and broadcasting a game event.
// It ensures a consistent order of operations:
// 1. An event is created from a payload.
// 2. The phaseManager is notified, allowing the current phase to react and potentially trigger a transition.
// 3. The event is broadcast to external listeners and recorded in the game's history.
// eventType: The type of the game event.
// payload: The payload for the event, which must be a protobuf message.
func (em *eventManager) createAndProcess(eventType model.GameEventType, payload *model.EventPayload) {
	event := &model.GameEvent{
		Type:      eventType,
		Timestamp: timestamppb.Now(),
		Payload:   payload,
	}

	// Step 1: Apply the event to the game state through the appropriate handler.
	if err := em.applyEvent(event); err != nil {
		em.logger.Error("failed to apply event", zap.String("event", event.Type.String()), zap.Error(err))
		return
	}

	// Step 2: Let the current phase react to the event.
	em.engine.pm.handleEvent(event)

	// Step 3: Publish the event to external listeners and for logging.
	select {
	case em.dispatchGameEvent <- event:
		// Also log the event in the game state for player review
		em.engine.GameState.DayEvents = append(em.engine.GameState.DayEvents, event)
		em.engine.GameState.LoopEvents = append(em.engine.GameState.LoopEvents, event)
	default:
		em.logger.Warn("Game event channel full, dropping event", zap.String("eventType", event.Type.String()))
	}

	// TODO: Re-implement trigger logic here, after the state is fully updated.
	// em.engine.checkForTriggers(event)
}

func (em *eventManager) applyEvent(event *model.GameEvent) error {
	handler, ok := eventhandler.GetHandler(event.Type)
	if !ok {
		// Not all events have handlers, so this is not an error.
		// It simply means no state change is associated with this event by default.
		return nil
	}

	return handler.Handle(em.engine, event)
}

// eventsChannel returns the outbound channel for game events.
// Returns: A read-only channel of game events.
func (em *eventManager) eventsChannel() <-chan *model.GameEvent {
	return em.dispatchGameEvent
}

// close closes the event dispatch channel.
func (em *eventManager) close() {
	close(em.dispatchGameEvent)
}
