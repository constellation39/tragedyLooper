package engine

import (
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	model "tragedylooper/internal/game/proto/v1"
)

func (ge *GameEngine) GetCharacter(id int32) (*model.Character, bool) {
	char, ok := ge.GameState.Characters[id]
	return char, ok
}

func (ge *GameEngine) SetCharacterLocation(id int32, location model.LocationType) {
	if char, ok := ge.GameState.Characters[id]; ok {
		char.CurrentLocation = location
		ge.logger.Info("Character moved", zap.Int32("characterID", id), zap.String("location", string(location)))
		ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_CHARACTER_MOVED, &model.CharacterMovedEvent{CharacterId: id, NewLocation: location})
	}
}

func (ge *GameEngine) AdjustCharacterParanoia(id int32, amount int32) int32 {
	if char, ok := ge.GameState.Characters[id]; ok {
		char.Paranoia += amount
		ge.logger.Info("Character paranoia adjusted", zap.Int32("characterID", id), zap.Int32("amount", amount), zap.Int32("newParanoia", char.Paranoia))
		ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_PARANOIA_ADJUSTED, &model.ParanoiaAdjustedEvent{CharacterId: id, Amount: amount, NewParanoia: char.Paranoia})
		return char.Paranoia
	}
	return 0
}

func (ge *GameEngine) AdjustCharacterGoodwill(id int32, amount int32) int32 {
	if char, ok := ge.GameState.Characters[id]; ok {
		char.Goodwill += amount
		ge.logger.Info("Character goodwill adjusted", zap.Int32("characterID", id), zap.Int32("amount", amount), zap.Int32("newGoodwill", char.Goodwill))
		ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_GOODWILL_ADJUSTED, &model.GoodwillAdjustedEvent{CharacterId: id, Amount: amount, NewGoodwill: char.Goodwill})
		return char.Goodwill
	}
	return 0
}

func (ge *GameEngine) AdjustCharacterIntrigue(id int32, amount int32) int32 {
	if char, ok := ge.GameState.Characters[id]; ok {
		char.Intrigue += amount
		ge.logger.Info("Character intrigue adjusted", zap.Int32("characterID", id), zap.Int32("amount", amount), zap.Int32("newIntrigue", char.Intrigue))
		ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_INTRIGUE_ADJUSTED, &model.IntrigueAdjustedEvent{CharacterId: id, Amount: amount, NewIntrigue: char.Intrigue})
		return char.Intrigue
	}
	return 0
}

func (ge *GameEngine) PublishEvent(eventType model.GameEventType, payload proto.Message) {
	ge.publishGameEvent(eventType, payload)
}

// checkTragedyConditions checks if the conditions for a given tragedy are met.
func (ge *GameEngine) checkTragedyConditions(tragedy *model.TragedyCondition) bool {
	for _, cond := range tragedy.Conditions {
		char, ok := ge.GameState.Characters[cond.CharacterId]
		if !ok || !char.IsAlive {
			return false // Character not found or not alive
		}
		if char.CurrentLocation != cond.Location {
			return false // Location mismatch
		}
		if char.Paranoia < cond.MinParanoia {
			return false // Paranoia too low
		}
		if cond.IsAlone {
			countAtLocation := 0
			for _, otherChar := range ge.GameState.Characters {
				if otherChar.CurrentLocation == cond.Location && otherChar.IsAlive {
					countAtLocation++
				}
			}
			if countAtLocation > 1 {
				return false // Not alone
			}
		}
	}
	return true // All conditions met
}

// resetLoop resets the game state for a new loop.
func (ge *GameEngine) resetLoop() {
	ge.logger.Info("Resetting for new loop...")
	// Reset characters to their initial script configuration
	for _, charConfig := range ge.GameState.Script.Characters {
		if char, ok := ge.GameState.Characters[charConfig.Id]; ok {
			char.CurrentLocation = charConfig.InitialLocation
			char.Paranoia = 0
			char.Goodwill = 0
			char.Intrigue = 0
			char.IsAlive = true
			for i := range char.Abilities {
				char.Abilities[i].UsedThisLoop = false
			}
		}
	}
	// Reset card usage status for all players
	for _, player := range ge.GameState.Players {
		for i := range player.Hand {
			player.Hand[i].UsedThisLoop = false
		}
	}

	// Clear loop-specific state
	ge.GameState.PreventedTragedies = make(map[int32]bool)
	ge.GameState.PlayedCardsThisDay = make(map[int32]*model.CardList)
	ge.GameState.PlayedCardsThisLoop = make(map[int32]*model.CardList)
	ge.GameState.DayEvents = []*model.GameEvent{}
	ge.GameState.LoopEvents = []*model.GameEvent{}

	ge.logger.Info("Loop reset complete.")
}

// endGame transitions the game to a finished state.
func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.logger.Info("Game Over!", zap.String("winner", string(winner)))
	ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_GAME_OVER
	ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_GAME_OVER, &model.GameOverEvent{Winner: winner})
	ge.StopGameLoop()
}

// resetPlayerReadiness resets the ready status for all players.
func (ge *GameEngine) resetPlayerReadiness() {
	for playerID := range ge.playerReady {
		ge.playerReady[playerID] = false
	}
}
