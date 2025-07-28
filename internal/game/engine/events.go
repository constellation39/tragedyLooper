package engine

import (
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	model "tragedylooper/internal/game/proto/v1"
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
	case ge.gameEventChan <- event:
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
	case model.GameEventType_CHARACTER_MOVED:
		var e model.CharacterMovedEvent
		if err := event.Payload.UnmarshalTo(&e); err != nil {
			ge.logger.Error("Failed to unmarshal CharacterMovedEvent", zap.Error(err))
			return
		}
		if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
			char.CurrentLocation = e.NewLocation
		}
	case model.GameEventType_PARANOIA_ADJUSTED:
		var e model.ParanoiaAdjustedEvent
		if err := event.Payload.UnmarshalTo(&e); err != nil {
			ge.logger.Error("Failed to unmarshal ParanoiaAdjustedEvent", zap.Error(err))
			return
		}
		if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
			char.Paranoia = e.NewParanoia
		}
	case model.GameEventType_GOODWILL_ADJUSTED:
		var e model.GoodwillAdjustedEvent
		if err := event.Payload.UnmarshalTo(&e); err != nil {
			ge.logger.Error("Failed to unmarshal GoodwillAdjustedEvent", zap.Error(err))
			return
		}
		if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
			char.Goodwill = e.NewGoodwill
		}
	case model.GameEventType_INTRIGUE_ADJUSTED:
		var e model.IntrigueAdjustedEvent
		if err := event.Payload.UnmarshalTo(&e); err != nil {
			ge.logger.Error("Failed to unmarshal IntrigueAdjustedEvent", zap.Error(err))
			return
		}
		if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
			char.Intrigue = e.NewIntrigue
		}
	case model.GameEventType_TRAIT_ADDED:
		var e model.TraitAddedEvent
		if err := event.Payload.UnmarshalTo(&e); err != nil {
			ge.logger.Error("Failed to unmarshal TraitAddedEvent", zap.Error(err))
			return
		}
		if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
			// Avoid duplicates
			for _, t := range char.Traits {
				if t == e.Trait {
					return
				}
			}
			char.Traits = append(char.Traits, e.Trait)
		}
	case model.GameEventType_TRAIT_REMOVED:
		var e model.TraitRemovedEvent
		if err := event.Payload.UnmarshalTo(&e); err != nil {
			ge.logger.Error("Failed to unmarshal TraitRemovedEvent", zap.Error(err))
			return
		}
		if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
			for i, t := range char.Traits {
				if t == e.Trait {
					char.Traits = append(char.Traits[:i], char.Traits[i+1:]...)
					return
				}
			}
		}
	default:
		ge.logger.Warn("Unknown event type for processing", zap.String("eventType", event.Type.String()))
	}
}
