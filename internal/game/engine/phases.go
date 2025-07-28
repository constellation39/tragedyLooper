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
	ge.GameState.PlayedCardsThisDay = make(map[int32]*model.Card) // Clear cards for the new day

	ge.checkAndTriggerAbilities(model.TriggerType_ON_DAY_START, nil)

	ge.GameState.CurrentPhase = model.GamePhase_CARD_PLAY
	ge.publishGameEvent(model.GameEventType_DAY_ADVANCED, &model.DayAdvancedEvent{Day: ge.GameState.CurrentDay, Loop: ge.GameState.CurrentLoop})
}

func (ge *GameEngine) handleCardPlayPhase() {
	ge.checkAndTriggerAbilities(model.TriggerType_ON_CARD_PLAY_PHASE, nil)

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
		ge.GameState.CurrentPhase = model.GamePhase_CARD_REVEAL
	}
}

func (ge *GameEngine) handleCardRevealPhase() {
	ge.logger.Info("Card Reveal Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_CARD_REVEAL_PHASE, nil)
	// ge.publishGameEvent(model.GameEventType_CARD_PLAYED, &model.CardPlayedEvent{PlayedCards: ge.GameState.PlayedCardsThisDay})
	ge.GameState.CurrentPhase = model.GamePhase_CARD_RESOLVE
}

func (ge *GameEngine) handleCardResolvePhase() {
	ge.logger.Info("Card Resolve Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_CARD_RESOLVE_PHASE, nil)

	// Resolve cards in a specific order if necessary (e.g., by initiative). For now, iterate over players.
	for playerID, card := range ge.GameState.PlayedCardsThisDay {
		// The card itself contains the target information, set during the play action.
		// We construct a payload for the effect system.
		var payload *model.UseAbilityPayload
		switch t := card.Target.(type) {
		case *model.Card_TargetCharacterId:
			payload = &model.UseAbilityPayload{
				Target: &model.UseAbilityPayload_TargetCharacterId{TargetCharacterId: t.TargetCharacterId},
			}
		case *model.Card_TargetLocation:
			payload = &model.UseAbilityPayload{
				Target: &model.UseAbilityPayload_TargetLocation{TargetLocation: t.TargetLocation},
			}
		}

		if err := ge.applyEffect(card.Effect, nil, payload); err != nil {
			ge.logger.Error("Error applying card effect",
				zap.Error(err),
				zap.String("playerID", fmt.Sprint(playerID)),
				zap.String("cardName", card.Name),
			)
		}
		ge.GameState.CurrentPhase = model.GamePhase_ABILITIES
	}
}

func (ge *GameEngine) handleAbilitiesPhase() {
	ge.logger.Info("Abilities Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_ABILITY_PHASE, nil)

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
		ge.GameState.CurrentPhase = model.GamePhase_INCIDENTS
	}
}

func (ge *GameEngine) handleIncidentsPhase() {
	ge.logger.Info("Incidents Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_INCIDENT_PHASE, nil)

	// Check for incidents on the current day
	for _, incident := range ge.gameData.GetIncidents() {
		if incident.Day == ge.GameState.CurrentDay {
			ge.logger.Info("Incident triggered!", zap.String("incident_name", incident.Name))
			ge.publishGameEvent(model.GameEventType_INCIDENT_TRIGGERED, &model.IncidentTriggeredEvent{Incident: incident})
		}
	}

	tragedyOccurred := false
	for _, tragedy := range ge.GameState.Script.Tragedies {
		// Check if the tragedy is active for the day and hasn't been prevented
		if tragedy.Day == ge.GameState.CurrentDay && ge.GameState.ActiveTragedies[int32(tragedy.TragedyType)] && !ge.GameState.PreventedTragedies[int32(tragedy.TragedyType)] {
			if ge.checkConditions(tragedy.Conditions) {
				ge.logger.Info("Tragedy triggered!", zap.String("tragedy_type", string(tragedy.TragedyType)))
				ge.publishGameEvent(model.GameEventType_TRAGEDY_TRIGGERED, &model.TragedyTriggeredEvent{TragedyType: tragedy.TragedyType})
				tragedyOccurred = true
				ge.GameState.ActiveTragedies[int32(tragedy.TragedyType)] = false // A tragedy can only occur once
				break                                                            // Only one tragedy per day
			}
		}
	}

	if gameOver, winner := ge.checkGameEndConditions(); gameOver {
		ge.endGame(winner)
		return
	}

	if tragedyOccurred {
		ge.GameState.CurrentPhase = model.GamePhase_LOOP_END
	} else {
		ge.GameState.CurrentPhase = model.GamePhase_DAY_END
	}
}

func (ge *GameEngine) handleDayEndPhase() {
	ge.logger.Info("Day End Phase", zap.Int("day", int(ge.GameState.CurrentDay)))
	ge.GameState.CurrentDay++
	if ge.GameState.CurrentDay > ge.GameState.Script.DaysPerLoop {
		ge.GameState.CurrentPhase = model.GamePhase_LOOP_END
	} else {
		ge.GameState.CurrentPhase = model.GamePhase_SETUP
	}
}

func (ge *GameEngine) handleLoopEndPhase() {
	ge.logger.Info("Loop End Phase", zap.Int("loop", int(ge.GameState.CurrentLoop)))
	ge.checkAndTriggerAbilities(model.TriggerType_ON_LOOP_START, nil)

	if gameOver, winner := ge.checkGameEndConditions(); gameOver {
		ge.endGame(winner)
		return
	}

	// If it's the last loop, the outcome is final.
	if ge.GameState.CurrentLoop >= ge.GameState.Script.LoopCount {
		// If no one has won by the final loop, Protagonists win by default
		ge.endGame(model.PlayerRole_PROTAGONIST)
		return
	}

	// If no tragedy occurred and more loops are left, reset for the next loop.
	ge.resetLoop()
	ge.GameState.CurrentLoop++
	ge.GameState.CurrentDay = 1
	ge.GameState.CurrentPhase = model.GamePhase_SETUP
	ge.publishGameEvent(model.GameEventType_LOOP_RESET, &model.LoopResetEvent{Loop: ge.GameState.CurrentLoop})
}

func (ge *GameEngine) handleProtagonistGuessPhase() {
	ge.logger.Info("Protagonist Guess Phase")
	// This phase is triggered by a player action (ActionMakeGuess).
	// The logic is handled in `handleMakeGuessAction`.
	// After the guess, the game transitions to GameOver.
	ge.GameState.CurrentPhase = model.GamePhase_GAME_OVER
}
