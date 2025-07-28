package engine

import (
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	model "tragedylooper/internal/game/proto/v1"
)

func (ge *GameEngine) applyAndPublishEvent(eventType model.GameEventType, payload proto.Message) {
	anyPayload, err := anypb.New(payload)
	if err != nil {
		ge.logger.Error("Failed to create anypb.Any for event payload", zap.Error(err))
		return
	}
	event := &model.GameEvent{
		Type:      eventType,
		Payload:   anyPayload,
		Timestamp: timestamppb.Now(),
	}

	// First, process the event to apply state changes synchronously
	ge.processEvent(event)

	// Then, publish the event for external listeners
	ge.publishGameEvent(event)
}

func (ge *GameEngine) publishGameEvent(event *model.GameEvent) {
	select {
	case ge.dispatchGameEvent <- event:
		// Also record the event in the game state for player views
		ge.GameState.DayEvents = append(ge.GameState.DayEvents, event)
		ge.GameState.LoopEvents = append(ge.GameState.LoopEvents, event)
	default:
		ge.logger.Warn("Game event channel full, dropping event", zap.String("eventType", event.Type.String()))
	}
}

func (ge *GameEngine) processEvent(event *model.GameEvent) {
	if handler, ok := eventHandlers[event.Type]; ok {
		handler(ge, event)
	} else {
		ge.logger.Warn("Unknown event type for processing", zap.String("eventType", event.Type.String()))
	}
}
