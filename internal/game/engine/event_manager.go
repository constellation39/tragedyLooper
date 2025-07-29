package engine

import (
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
func (em *eventManager) createAndProcess(eventType model.GameEventType, payload proto.Message) {
	event := &model.GameEvent{
		Type:      eventType,
		Timestamp: timestamppb.Now(),
	}

	switch p := payload.(type) {
	case *model.CharacterMovedEvent:
		event.Payload = &model.GameEvent_CharacterMoved{CharacterMoved: p}
	case *model.ParanoiaAdjustedEvent:
		event.Payload = &model.GameEvent_ParanoiaAdjusted{ParanoiaAdjusted: p}
	case *model.GoodwillAdjustedEvent:
		event.Payload = &model.GameEvent_GoodwillAdjusted{GoodwillAdjusted: p}
	case *model.IntrigueAdjustedEvent:
		event.Payload = &model.GameEvent_IntrigueAdjusted{IntrigueAdjusted: p}
	case *model.LoopLossEvent:
		event.Payload = &model.GameEvent_LoopLoss{LoopLoss: p}
	case *model.LoopWinEvent:
		event.Payload = &model.GameEvent_LoopWin{LoopWin: p}
	case *model.AbilityUsedEvent:
		event.Payload = &model.GameEvent_AbilityUsed{AbilityUsed: p}
	case *model.DayAdvancedEvent:
		event.Payload = &model.GameEvent_DayAdvanced{DayAdvanced: p}
	case *model.CardPlayedEvent:
		event.Payload = &model.GameEvent_CardPlayed{CardPlayed: p}
	case *model.CardRevealedEvent:
		event.Payload = &model.GameEvent_CardRevealed{CardRevealed: p}
	case *model.LoopResetEvent:
		event.Payload = &model.GameEvent_LoopReset{LoopReset: p}
	case *model.GameOverEvent:
		event.Payload = &model.GameEvent_GameOver{GameOver: p}
	case *model.ChoiceRequiredEvent:
		event.Payload = &model.GameEvent_ChoiceRequired{ChoiceRequired: p}
	case *model.IncidentTriggeredEvent:
		event.Payload = &model.GameEvent_IncidentTriggered{IncidentTriggered: p}
	case *model.TragedyTriggeredEvent:
		event.Payload = &model.GameEvent_TragedyTriggered{TragedyTriggered: p}
	case *model.TraitAddedEvent:
		event.Payload = &model.GameEvent_TraitAdded{TraitAdded: p}
	case *model.TraitRemovedEvent:
		event.Payload = &model.GameEvent_TraitRemoved{TraitRemoved: p}
	case *model.PlayerActionTakenEvent:
		event.Payload = &model.GameEvent_PlayerActionTaken{PlayerActionTaken: p}
	default:
		em.logger.Error("Unknown event payload type", zap.String("type", string(p.ProtoReflect().Descriptor().FullName())))
		return
	}

	// Step 1: Let the current phase react to the event.
	// This is now handled by the phase manager.
	em.engine.pm.handleEvent(event)

	// Step 2: Publish the event to external listeners and for logging.
	// This happens after the state has been updated.
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

// eventsChannel returns the outbound channel for game events.
// Returns: A read-only channel of game events.
func (em *eventManager) eventsChannel() <-chan *model.GameEvent {
	return em.dispatchGameEvent
}

// close closes the event dispatch channel.
func (em *eventManager) close() {
	close(em.dispatchGameEvent)
}
