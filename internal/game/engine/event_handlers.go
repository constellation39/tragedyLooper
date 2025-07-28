package engine

import (
	"go.uber.org/zap"

	model "tragedylooper/internal/game/proto/v1"
)

// eventHandler is a function that handles a specific game event.
type eventHandler func(ge *GameEngine, event *model.GameEvent)

// eventHandlers maps event types to their respective handler functions.
var eventHandlers = map[model.GameEventType]eventHandler{
	model.GameEventType_CHARACTER_MOVED:    handleCharacterMoved,
	model.GameEventType_PARANOIA_ADJUSTED:  handleParanoiaAdjusted,
	model.GameEventType_GOODWILL_ADJUSTED:  handleGoodwillAdjusted,
	model.GameEventType_INTRIGUE_ADJUSTED:  handleIntrigueAdjusted,
	model.GameEventType_TRAIT_ADDED:        handleTraitAdded,
	model.GameEventType_TRAIT_REMOVED:      handleTraitRemoved,
	model.GameEventType_CARD_PLAYED:        handleCardPlayed,
	model.GameEventType_CARD_REVEALED:      handleCardRevealed,
	model.GameEventType_DAY_ADVANCED:       handleDayAdvanced,
	model.GameEventType_LOOP_RESET:         handleLoopReset,
	model.GameEventType_GAME_ENDED:         handleGameOver,
	model.GameEventType_INCIDENT_TRIGGERED: handleIncidentTriggered,
	model.GameEventType_LOOP_WIN:           handleLoopWin,
	model.GameEventType_LOOP_LOSS:          handleLoopLoss,
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

func handleCardPlayed(ge *GameEngine, event *model.GameEvent) {
	var e model.CardPlayedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal CardPlayedEvent", zap.Error(err))
		return
	}
	ge.logger.Info("Card played", zap.Int32("player_id", e.PlayerId), zap.String("card_name", e.Card.Config.Name))
}

func handleCardRevealed(ge *GameEngine, event *model.GameEvent) {
	var e model.CardRevealedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal CardRevealedEvent", zap.Error(err))
		return
	}
	cardNames := make([]string, len(e.Cards))
	for i, c := range e.Cards {
		cardNames[i] = c.Config.Name
	}
	ge.logger.Info("Cards revealed", zap.Strings("cards", cardNames))
}

func handleDayAdvanced(ge *GameEngine, event *model.GameEvent) {
	var e model.DayAdvancedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal DayAdvancedEvent", zap.Error(err))
		return
	}
	ge.logger.Info("Day advanced", zap.Int32("day", e.Day), zap.Int32("loop", e.Loop))
	ge.GameState.DayEvents = []*model.GameEvent{}
}

func handleLoopReset(ge *GameEngine, event *model.GameEvent) {
	var e model.LoopResetEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal LoopResetEvent", zap.Error(err))
		return
	}
	ge.logger.Info("Loop reset", zap.Int32("loop", e.Loop))
	ge.GameState.LoopEvents = []*model.GameEvent{}
}

func handleGameOver(ge *GameEngine, event *model.GameEvent) {
	var e model.GameOverEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal GameOverEvent", zap.Error(err))
		return
	}
	ge.logger.Info("Game over", zap.String("winner", e.Winner.String()))
}

func handleIncidentTriggered(ge *GameEngine, event *model.GameEvent) {
	var e model.IncidentTriggeredEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal IncidentTriggeredEvent", zap.Error(err))
		return
	}
	ge.logger.Info("Incident triggered", zap.String("incident_name", e.Incident.Name))
}

func handleLoopWin(ge *GameEngine, event *model.GameEvent) {
	var e model.LoopWinEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal LoopWinEvent", zap.Error(err))
		return
	}
	ge.logger.Info("Loop win", zap.Int32("loop", ge.GameState.CurrentLoop))
}

func handleLoopLoss(ge *GameEngine, event *model.GameEvent) {
	var e model.LoopLossEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		ge.logger.Error("Failed to unmarshal LoopLossEvent", zap.Error(err))
		return
	}
	ge.logger.Info("Loop loss", zap.Int32("loop", ge.GameState.CurrentLoop))
}
