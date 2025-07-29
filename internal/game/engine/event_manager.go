package engine

import (
	"tragedylooper/internal/game/engine/eventhandler"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// eventManager is responsible for creating, processing, and dispatching all game events.
// It decouples the event lifecycle from the main GameEngine, ensuring a clear and maintainable flow.
type eventManager struct {
	engine *GameEngine // Reference to the parent engine to access global state and managers.
	logger *zap.Logger

	// dispatchGameEvent is an outbound channel that broadcasts processed game events to external listeners.
	dispatchGameEvent chan *model.GameEvent
}

func newEventManager(engine *GameEngine) *eventManager {
	return &eventManager{
		engine:            engine,
		logger:            engine.logger.Named("EventManager"),
		dispatchGameEvent: make(chan *model.GameEvent, 100),
	}
}

// createAndProcess is the central method for creating, applying, and broadcasting game events.
// It ensures a consistent order of operations:
// 1. The event is created from a payload.
// 2. The game state is mutated synchronously by the event handler.
// 3. The phaseManager is notified, allowing the current phase to react and potentially trigger a transition.
// 4. The event is broadcast to external listeners and recorded in the game's history.
func (em *eventManager) createAndProcess(eventType model.GameEventType, payload proto.Message) {
	anyPayload, err := anypb.New(payload)
	if err != nil {
		em.logger.Error("Failed to create anypb.Any for event payload", zap.Error(err))
		return
	}
	event := &model.GameEvent{
		Type:      eventType,
		Payload:   anyPayload,
		Timestamp: timestamppb.Now(),
	}

	// Step 1: Apply the state change synchronously.
	// This is critical to ensure the game state is consistent before any other logic runs.
	if err := eventhandler.ProcessEvent(em.engine.GameState, event); err != nil {
		em.logger.Error("Failed to apply event to game state", zap.Error(err), zap.String("type", event.Type.String()))
		// We continue even if the handler fails, to allow the phase and listeners to react.
	}

	// Step 2: Let the current phase react to the event.
	// This is now handled by the phase manager.
	em.engine.pm.handleEvent(event)

	// Step 3: Publish the event to external listeners and record it.
	// This happens after the state has been updated.
	select {
	case em.dispatchGameEvent <- event:
		// Also record the event in the game state for player views
		em.engine.GameState.DayEvents = append(em.engine.GameState.DayEvents, event)
		em.engine.GameState.LoopEvents = append(em.engine.GameState.LoopEvents, event)
	default:
		em.logger.Warn("Game event channel full, dropping event", zap.String("eventType", event.Type.String()))
	}

	// TODO: Re-implement trigger logic here, after state has fully updated.
	// em.engine.checkForTriggers(event)
}

// eventsChannel returns the outbound channel for game events.
func (em *eventManager) eventsChannel() <-chan *model.GameEvent {
	return em.dispatchGameEvent
}

func (em *eventManager) close() {
	close(em.dispatchGameEvent)
}
