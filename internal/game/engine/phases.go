package engine

import (
	"fmt"
	model "tragedylooper/internal/game/proto/v1"

	"go.uber.org/zap"
)

// --- Game Phase Handlers ---

func (ge *GameEngine) handleMorningPhase() {
	ge.logger.Info("Morning Phase", zap.Int("loop", int(ge.GameState.CurrentLoop)), zap.Int("day", int(ge.GameState.CurrentDay)))
	ge.resetPlayerReadiness()
	ge.GameState.PlayedCardsThisDay = make(map[int32]*model.CardList) // Clear cards for the new day

	// Trigger DayStart abilities
	for _, char := range ge.GameState.Characters {
		for i, ability := range char.Abilities {
			if ability.TriggerType == model.AbilityTriggerType_ABILITY_TRIGGER_TYPE_DAY_START && !ability.UsedThisLoop {
				payload := model.UseAbilityPayload{CharacterId: char.Id, AbilityId: ability.Id} // Assuming self-target for simplicity
				if err := ge.applyEffect(ability.Effect, ability, &payload); err != nil {
					ge.logger.Error("Error applying DayStart ability effect", zap.Error(err), zap.String("character", char.Name), zap.String("ability", ability.Name))
				}
				ge.GameState.Characters[char.Id].Abilities[i].UsedThisLoop = true // Mark as used
				ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_ABILITY_USED, &model.AbilityUsedEvent{CharacterId: char.Id, AbilityName: ability.Name})
			}
		}
	}

	ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_CARD_PLAY
	ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_DAY_ADVANCED, &model.DayAdvancedEvent{Day: ge.GameState.CurrentDay, Loop: ge.GameState.CurrentLoop})
}

func (ge *GameEngine) handleCardPlayPhase() {
	allPlayersReady := true
	for playerID, player := range ge.GameState.Players {
		if ge.playerReady[playerID] {
			continue
		}
		if player.IsLlm {
			go ge.triggerLLMPlayerAction(playerID)
		}
		allPlayersReady = false
	}

	if allPlayersReady {
		ge.logger.Info("All players ready for Card Reveal.")
		ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_CARD_REVEAL
	}
}

func (ge *GameEngine) handleCardRevealPhase() {
	ge.logger.Info("Card Reveal Phase")
	ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_CARD_PLAYED, &model.CardPlayedEvent{PlayedCards: ge.GameState.PlayedCardsThisDay})
	ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_CARD_RESOLVE
}

func (ge *GameEngine) handleCardResolvePhase() {
	ge.logger.Info("Card Resolve Phase")
	// Resolve cards in a specific order if necessary (e.g., by initiative). For now, iterate over players.
	for playerID, cards := range ge.GameState.PlayedCardsThisDay {
		for _, card := range cards.Cards {
			// The card itself contains the target information, set during the play action.
			// We construct a payload for the effect system.
			payload := &model.UseAbilityPayload{
				Target: card.Target,
			}
			if err := ge.applyEffect(card.Effect, nil, payload); err != nil {
				ge.logger.Error("Error applying card effect",
					zap.Error(err),
					zap.String("playerID", fmt.Sprint(playerID)),
					zap.String("cardName", card.Name),
				)
			}
		}
	}
	ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_ABILITIES
}

func (ge *GameEngine) handleAbilitiesPhase() {
	ge.logger.Info("Abilities Phase")

	allPlayersReady := true
	for playerID, player := range ge.GameState.Players {
		if ge.playerReady[playerID] {
			continue
		}
		if player.IsLlm {
			go ge.triggerLLMPlayerAction(playerID)
		}
		allPlayersReady = false
	}

	if allPlayersReady {
		ge.logger.Info("All players ready for Incidents Phase.")
		ge.resetPlayerReadiness()
		ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_INCIDENTS
	}
}

func (ge *GameEngine) handleIncidentsPhase() {
	ge.logger.Info("Incidents Phase")

	// Check for incidents on the current day
	for _, incident := range ge.GameState.Script.Incidents {
		if incident.Day == ge.GameState.CurrentDay {
			ge.logger.Info("Incident triggered!", zap.String("incident_name", incident.Name))
			ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_INCIDENT_TRIGGERED, &model.IncidentTriggeredEvent{Incident: incident})
		}
	}

	tragedyOccurred := false
	for _, tragedy := range ge.GameState.Script.Tragedies {
		// Check if the tragedy is active for the day and hasn't been prevented
		if tragedy.Day == ge.GameState.CurrentDay && ge.GameState.ActiveTragedies[int32(tragedy.TragedyType)] && !ge.GameState.PreventedTragedies[int32(tragedy.TragedyType)] {
			if ge.checkConditions(tragedy.Conditions) {
				ge.logger.Info("Tragedy triggered!", zap.String("tragedy_type", string(tragedy.TragedyType)))
				ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_TRAGEDY_TRIGGERED, &model.TragedyTriggeredEvent{TragedyType: tragedy.TragedyType})
				tragedyOccurred = true
				ge.GameState.TragedyOccurred[int32(tragedy.TragedyType)] = true
				break // Only one tragedy per day
			}
		}
	}

	if gameOver, winner := ge.checkGameEndConditions(); gameOver {
		ge.endGame(winner)
		return
	}

	if tragedyOccurred {
		ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_LOOP_END
	} else {
		ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_DAY_END
	}
}

func (ge *GameEngine) handleDayEndPhase() {
	ge.logger.Info("Day End Phase", zap.Int("day", int(ge.GameState.CurrentDay)))
	ge.GameState.CurrentDay++
	if ge.GameState.CurrentDay > ge.GameState.Script.DaysPerLoop {
		ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_LOOP_END
	} else {
		ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_MORNING
	}
}

func (ge *GameEngine) handleLoopEndPhase() {
	ge.logger.Info("Loop End Phase", zap.Int("loop", int(ge.GameState.CurrentLoop)))

	if gameOver, winner := ge.checkGameEndConditions(); gameOver {
		ge.endGame(winner)
		return
	}

	// If it's the last loop, the outcome is final.
	if ge.GameState.CurrentLoop >= ge.GameState.Script.LoopCount {
		// If no one has won by the final loop, Protagonists win by default
		ge.endGame(model.PlayerRole_PLAYER_ROLE_PROTAGONIST)
		return
	}

	// If no tragedy occurred and more loops are left, reset for the next loop.
	ge.resetLoop()
	ge.GameState.CurrentLoop++
	ge.GameState.CurrentDay = 1
	ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_MORNING
	ge.publishGameEvent(model.GameEventType_GAME_EVENT_TYPE_LOOP_RESET, &model.LoopResetEvent{Loop: ge.GameState.CurrentLoop})
}

func (ge *GameEngine) handleProtagonistGuessPhase() {
	ge.logger.Info("Protagonist Guess Phase")
	// This phase is triggered by a player action (ActionMakeGuess).
	// The logic is handled in `handleMakeGuessAction`.
	// After the guess, the game transitions to GameOver.
	ge.GameState.CurrentPhase = model.GamePhase_GAME_PHASE_GAME_OVER
}
