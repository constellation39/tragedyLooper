package engine

import (
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"tragedylooper/internal/game/proto/model"
)

func (ge *GameEngine) publishGameEvent(eventType model.GameEventType, payload proto.Message) {
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
	select {
	case ge.gameEventChan <- *event:
		// Also record the event in the game state for player views
		ge.GameState.DayEvents = append(ge.GameState.DayEvents, event)
		ge.GameState.LoopEvents = append(ge.GameState.LoopEvents, event)
	default:
		ge.logger.Warn("Game event channel full, dropping event", zap.String("eventType", string(eventType)))
	}
}

func (ge *GameEngine) processEvent(event *model.GameEvent) {
	// This function applies the consequences of a resolved effect event to the game state.
	switch event.Type {
	case model.GameEventType_GAME_EVENT_TYPE_CHARACTER_MOVED:
		var e model.CharacterMovedEvent
		if err := event.Payload.UnmarshalTo(&e); err != nil {
			ge.logger.Error("Failed to unmarshal CharacterMovedEvent", zap.Error(err))
			return
		}
		if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
			char.CurrentLocation = e.NewLocation
		}
	case model.GameEventType_GAME_EVENT_TYPE_PARANOIA_ADJUSTED:
		var e model.ParanoiaAdjustedEvent
		if err := event.Payload.UnmarshalTo(&e); err != nil {
			ge.logger.Error("Failed to unmarshal ParanoiaAdjustedEvent", zap.Error(err))
			return
		}
		if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
			char.Paranoia += e.Amount
		}
	case model.GameEventType_GAME_EVENT_TYPE_GOODWILL_ADJUSTED:
		var e model.GoodwillAdjustedEvent
		if err := event.Payload.UnmarshalTo(&e); err != nil {
			ge.logger.Error("Failed to unmarshal GoodwillAdjustedEvent", zap.Error(err))
			return
		}
		if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
			char.Goodwill += e.Amount
		}
	case model.GameEventType_GAME_EVENT_TYPE_INTRIGUE_ADJUSTED:
		var e model.IntrigueAdjustedEvent
		if err := event.Payload.UnmarshalTo(&e); err != nil {
			ge.logger.Error("Failed to unmarshal IntrigueAdjustedEvent", zap.Error(err))
			return
		}
		if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
			char.Intrigue += e.Amount
		}
	default:
		ge.logger.Warn("Unknown event type for processing", zap.String("eventType", event.Type.String()))
	}
}
