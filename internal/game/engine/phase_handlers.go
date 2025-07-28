package engine

import (
	model "tragedylooper/internal/game/proto/v1"

	"go.uber.org/zap"
)

// phaseHandler is a function that handles the logic for a specific game phase.
type phaseHandler func(ge *GameEngine)

// phaseHandlers maps game phases to their respective handler functions.
var phaseHandlers = map[model.GamePhase]phaseHandler{
	model.GamePhase_SETUP:             handleSetupPhase,
	model.GamePhase_MASTERMIND_SETUP:  handleMastermindSetupPhase,
	model.GamePhase_CARD_PLAY:         handleCardPlayPhase,
	model.GamePhase_CARD_REVEAL:       handleCardRevealPhase,
	model.GamePhase_CARD_RESOLVE:      handleCardResolvePhase,
	model.GamePhase_ABILITIES:         handleAbilitiesPhase,
	model.GamePhase_INCIDENTS:         handleIncidentsPhase,
	model.GamePhase_DAY_END:           handleDayEndPhase,
	model.GamePhase_LOOP_END:          handleLoopEndPhase,
	model.GamePhase_PROTAGONIST_GUESS: handleProtagonistGuessPhase,
	model.GamePhase_GAME_OVER:         handleGameOverPhase,
}

// handleSetupPhase is called at the beginning of each day.
func handleSetupPhase(ge *GameEngine) {
	ge.logger.Info("Morning Phase (Setup)", zap.Int("loop", int(ge.GameState.CurrentLoop)), zap.Int("day", int(ge.GameState.CurrentDay)))
	ge.resetPlayerReadiness()
	ge.GameState.PlayedCardsThisDay = make(map[int32]*model.Card) // Clear cards for the new day

	ge.checkAndTriggerAbilities(model.TriggerType_ON_DAY_START)

	ge.GameState.CurrentPhase = model.GamePhase_CARD_PLAY
	ge.applyAndPublishEvent(model.GameEventType_DAY_ADVANCED, &model.DayAdvancedEvent{Day: ge.GameState.CurrentDay, Loop: ge.GameState.CurrentLoop})
}

// handleMastermindSetupPhase allows the mastermind to perform setup actions at the beginning of a loop.
func handleMastermindSetupPhase(ge *GameEngine) {
	ge.logger.Info("Mastermind Setup Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)

	// In a real implementation, this phase would wait for the Mastermind player to perform specific setup actions.
	// For now, we assume it's an automatic phase transition if the player is ready.
	mastermind := ge.getMastermindPlayer()
	if mastermind == nil {
		ge.logger.Error("No mastermind player found during mastermind setup phase")
		ge.GameState.CurrentPhase = model.GamePhase_SETUP // Skip to next phase to avoid getting stuck
		return
	}

	// If the mastermind is an LLM, it might have automated setup logic.
	if mastermind.IsLlm && !ge.playerReady[mastermind.Id] {
		go ge.triggerLLMPlayerAction(mastermind.Id)
	}

	// Once the mastermind is ready, proceed.
	if ge.playerReady[mastermind.Id] {
		ge.logger.Info("Mastermind setup complete.")
		ge.resetPlayerReadiness()
		ge.GameState.CurrentPhase = model.GamePhase_SETUP
	}
	// Otherwise, we wait in this phase for the mastermind to become ready.
}

// handleCardPlayPhase waits for all players to play their cards.
func handleCardPlayPhase(ge *GameEngine) {
	ge.logger.Info("Card Play Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)

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
		ge.logger.Info("All players have played their cards.")
		ge.resetPlayerReadiness()
		ge.GameState.CurrentPhase = model.GamePhase_CARD_REVEAL
	}
}

// handleCardRevealPhase reveals the cards played this day.
func handleCardRevealPhase(ge *GameEngine) {
	ge.logger.Info("Card Reveal Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)

	// Create a list of played cards to include in the event
	playedCards := make([]*model.Card, 0, len(ge.GameState.PlayedCardsThisDay))
	for _, card := range ge.GameState.PlayedCardsThisDay {
		playedCards = append(playedCards, card)
	}

	// Publish an event with all the cards that were played this turn.
	ge.applyAndPublishEvent(model.GameEventType_CARD_REVEALED, &model.CardRevealedEvent{Cards: playedCards})

	ge.GameState.CurrentPhase = model.GamePhase_CARD_RESOLVE
}

// handleCardResolvePhase resolves the effects of all played cards.
func handleCardResolvePhase(ge *GameEngine) {
	ge.logger.Info("Card Resolve Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)

	// TODO: Resolve cards in a specific order (e.g., by priority).
	for playerID, card := range ge.GameState.PlayedCardsThisDay {
		ge.applyAndPublishEvent(model.GameEventType_CARD_PLAYED, &model.CardPlayedEvent{PlayerId: playerID, Card: card})

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

		effect := &model.Effect{EffectType: &model.Effect_CompoundEffect{CompoundEffect: card.Config.Effect}}
		if err := ge.applyEffect(effect, ge.findPlayer(playerID), payload, nil); err != nil {
			ge.logger.Error("Error applying card effect",
				zap.Error(err),
				zap.Int32("playerID", playerID),
				zap.String("cardName", card.Config.Name),
			)
		}
	}
	ge.GameState.CurrentPhase = model.GamePhase_ABILITIES
}

// handleAbilitiesPhase waits for players to use their abilities.
func handleAbilitiesPhase(ge *GameEngine) {
	ge.logger.Info("Abilities Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)

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
		ge.logger.Info("All players are done with the abilities phase.")
		ge.resetPlayerReadiness()
		ge.GameState.CurrentPhase = model.GamePhase_INCIDENTS
	}
}

// handleIncidentsPhase checks for and triggers any incidents for the current day.
func handleIncidentsPhase(ge *GameEngine) {
	ge.logger.Info("Incidents Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)

	script, err := ge.gameConfig.GetScript()
	if err != nil {
		ge.logger.Error("Failed to get script for incidents phase", zap.Error(err))
		ge.endGame(model.PlayerRole_PLAYER_ROLE_UNSPECIFIED) // End game on error
		return
	}

	// Check for incidents on the current day
	for _, incident := range script.Incidents {
		if incident.Day == ge.GameState.CurrentDay {
			if ge.checkConditions(incident.TriggerConditions, nil, nil, nil) {
				ge.logger.Info("Incident triggered!", zap.String("incident_name", incident.Name))
				effect := &model.Effect{EffectType: &model.Effect_CompoundEffect{CompoundEffect: incident.Effect}}
				if err := ge.applyEffect(effect, nil, nil, nil); err != nil {
					ge.logger.Error("failed to apply incident effect", zap.Error(err), zap.String("incident", incident.Name))
				}
				ge.applyAndPublishEvent(model.GameEventType_INCIDENT_TRIGGERED, &model.IncidentTriggeredEvent{Incident: &model.Incident{Config: incident, Name: incident.Name, Day: incident.Day}})
			}
		}
	}

	if gameOver, winner := ge.checkGameEndConditions(); gameOver {
		ge.endGame(winner)
		return
	}

	ge.GameState.CurrentPhase = model.GamePhase_DAY_END
}

// handleDayEndPhase advances the day or transitions to the loop end phase.
func handleDayEndPhase(ge *GameEngine) {
	ge.logger.Info("Day End Phase", zap.Int("day", int(ge.GameState.CurrentDay)))
	ge.checkAndTriggerAbilities(model.TriggerType_ON_DAY_END)

	ge.GameState.CurrentDay++
	script, err := ge.gameConfig.GetScript()
	if err != nil {
		ge.logger.Error("Failed to get script for day end phase", zap.Error(err))
		ge.endGame(model.PlayerRole_PLAYER_ROLE_UNSPECIFIED) // End game on error
		return
	}

	if ge.GameState.CurrentDay > script.DaysPerLoop {
		ge.GameState.CurrentPhase = model.GamePhase_LOOP_END
	} else {
		ge.GameState.CurrentPhase = model.GamePhase_SETUP
	}
}

// handleLoopEndPhase checks win/loss conditions for the loop and either starts a new loop or ends the game.
func handleLoopEndPhase(ge *GameEngine) {
	ge.logger.Info("Loop End Phase", zap.Int("loop", int(ge.GameState.CurrentLoop)))
	ge.checkAndTriggerAbilities(model.TriggerType_ON_LOOP_END)

	if gameOver, winner := ge.checkGameEndConditions(); gameOver {
		ge.endGame(winner)
		return
	}

	script, err := ge.gameConfig.GetScript()
	if err != nil {
		ge.logger.Error("Failed to get script for loop end phase", zap.Error(err))
		ge.endGame(model.PlayerRole_PLAYER_ROLE_UNSPECIFIED) // End game on error
		return
	}

	if ge.GameState.CurrentLoop >= script.LoopCount {
		ge.logger.Info("Final loop has ended.")
		ge.applyAndPublishEvent(model.GameEventType_LOOP_WIN, &model.LoopWinEvent{Loop: ge.GameState.CurrentLoop})
		ge.endGame(model.PlayerRole_PROTAGONIST)
		return
	}

	ge.applyAndPublishEvent(model.GameEventType_LOOP_LOSS, &model.LoopLossEvent{Loop: ge.GameState.CurrentLoop})
	ge.resetLoop()
	ge.GameState.CurrentLoop++
	ge.GameState.CurrentDay = 1
	ge.GameState.CurrentPhase = model.GamePhase_MASTERMIND_SETUP
	ge.applyAndPublishEvent(model.GameEventType_LOOP_RESET, &model.LoopResetEvent{Loop: ge.GameState.CurrentLoop})
	ge.checkAndTriggerAbilities(model.TriggerType_ON_LOOP_START)
}

// handleProtagonistGuessPhase waits for a protagonist to make a guess.
func handleProtagonistGuessPhase(ge *GameEngine) {
	ge.logger.Info("Protagonist Guess Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)

	// The engine now waits for an ActionMakeGuess from a protagonist player.
	// The logic for handling the guess is in `actions.go:handleMakeGuessAction`.
	for playerID, player := range ge.GameState.Players {
		if player.Role == model.PlayerRole_PROTAGONIST && !ge.playerReady[playerID] && player.IsLlm {
			go ge.triggerLLMPlayerAction(playerID)
		}
	}
	// We stay in this phase until a guess is made.
}

// handleGameOverPhase is the final phase of the game.
func handleGameOverPhase(ge *GameEngine) {
	ge.logger.Info("Game Over Phase. The game has concluded.")
	// This is a terminal phase. No further actions are taken.
}
