package engine

import (
	"go.uber.org/zap"

	model "tragedylooper/internal/game/proto/v1"
)

// eventHandler is a function that handles a specific game event.
type eventHandler func(ge *GameEngine, event *model.GameEvent)

// eventHandlers maps event types to their respective handler functions.
var eventHandlers = map[model.GameEventType]eventHandler{
	model.GameEventType_CHARACTER_MOVED:   handleCharacterMoved,
	model.GameEventType_PARANOIA_ADJUSTED: handleParanoiaAdjusted,
	model.GameEventType_GOODWILL_ADJUSTED: handleGoodwillAdjusted,
	model.GameEventType_INTRIGUE_ADJUSTED: handleIntrigueAdjusted,
	model.GameEventType_TRAIT_ADDED:       handleTraitAdded,
	model.GameEventType_TRAIT_REMOVED:     handleTraitRemoved,
}

func handleCharacterMoved(ge *GameEngine, event *model.GameEvent) {
	var e model.CharacterMovedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal CharacterMovedEvent", zap.Error(err))
		return
	}
	if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
		char.CurrentLocation = e.NewLocation
	}
}

func handleParanoiaAdjusted(ge *GameEngine, event *model.GameEvent) {
	var e model.ParanoiaAdjustedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal ParanoiaAdjustedEvent", zap.Error(err))
		return
	}
	if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
		char.Paranoia += e.Amount
		e.NewParanoia = char.Paranoia
	}
}

func handleGoodwillAdjusted(ge *GameEngine, event *model.GameEvent) {
	var e model.GoodwillAdjustedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal GoodwillAdjustedEvent", zap.Error(err))
		return
	}
	if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
		char.Goodwill += e.Amount
		e.NewGoodwill = char.Goodwill
	}
}

func handleIntrigueAdjusted(ge *GameEngine, event *model.GameEvent) {
	var e model.IntrigueAdjustedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal IntrigueAdjustedEvent", zap.Error(err))
		return
	}
	if char, ok := ge.GameState.Characters[e.CharacterId]; ok {
		char.Intrigue += e.Amount
		e.NewIntrigue = char.Intrigue
	}
}

func handleTraitAdded(ge *GameEngine, event *model.GameEvent) {
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
}

func handleTraitRemoved(ge *GameEngine, event *model.GameEvent) {
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
}
