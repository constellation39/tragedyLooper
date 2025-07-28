package engine

import (
	"fmt"
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
	model.GamePhase_GAME_OVER:         handleGameOverPhase,
	model.GamePhase_PROTAGONIST_GUESS: handleProtagonistGuessPhase,
}

func handleSetupPhase(ge *GameEngine) {
	ge.handleMorningPhase()
}

func handleMastermindSetupPhase(ge *GameEngine) {
	// Implementation for Mastermind Setup phase
}

func handleCardPlayPhase(ge *GameEngine) {
	ge.handleCardPlayPhase()
}

func handleCardRevealPhase(ge *GameEngine) {
	ge.handleCardRevealPhase()
}

func handleCardResolvePhase(ge *GameEngine) {
	ge.handleCardResolvePhase()
}

func handleAbilitiesPhase(ge *GameEngine) {
	ge.handleAbilitiesPhase()
}

func handleIncidentsPhase(ge *GameEngine) {
	ge.handleIncidentsPhase()
}

func handleDayEndPhase(ge *GameEngine) {
	ge.handleDayEndPhase()
}

func handleLoopEndPhase(ge *GameEngine) {
	ge.handleLoopEndPhase()
}

func handleGameOverPhase(ge *GameEngine) {
	ge.handleProtagonistGuessPhase()
}

func handleProtagonistGuessPhase(ge *GameEngine) {
	// Implementation for Protagonist Guess phase
}

func (ge *GameEngine) handleMorningPhase() {
	ge.logger.Info("Morning Phase", zap.Int("loop", int(ge.GameState.CurrentLoop)), zap.Int("day", int(ge.GameState.CurrentDay)))
	ge.resetPlayerReadiness()
	ge.GameState.PlayedCardsThisDay = make(map[int32]*model.Card) // Clear cards for the new day

	ge.checkAndTriggerAbilities(model.TriggerType_ON_DAY_START)

	ge.GameState.CurrentPhase = model.GamePhase_CARD_PLAY
	ge.applyAndPublishEvent(model.GameEventType_DAY_ADVANCED, &model.DayAdvancedEvent{Day: ge.GameState.CurrentDay, Loop: ge.GameState.CurrentLoop})
}

func (ge *GameEngine) handleCardPlayPhase() {
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
		ge.logger.Info("All players ready for Card Reveal.")
		ge.GameState.CurrentPhase = model.GamePhase_CARD_REVEAL
	}
}

func (ge *GameEngine) handleCardRevealPhase() {
	ge.logger.Info("Card Reveal Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)
	// ge.publishGameEvent(model.GameEventType_CARD_PLAYED, &model.CardPlayedEvent{PlayedCards: ge.GameState.PlayedCardsThisDay})
	ge.GameState.CurrentPhase = model.GamePhase_CARD_RESOLVE
}

func (ge *GameEngine) handleCardResolvePhase() {
	ge.logger.Info("Card Resolve Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)

	// Resolve cards in a specific order if necessary (e.g., by priority). For now, iterate over players.
	for playerID, card := range ge.GameState.PlayedCardsThisDay {
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
		if err := ge.applyEffect(effect, nil, payload, nil); err != nil {
			ge.logger.Error("Error applying card effect",
				zap.Error(err),
				zap.String("playerID", fmt.Sprint(playerID)),
				zap.String("cardName", card.Config.Name),
			)
		}
	}
	ge.GameState.CurrentPhase = model.GamePhase_ABILITIES
}

func (ge *GameEngine) handleAbilitiesPhase() {
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
		ge.logger.Info("All players ready for Incidents Phase.")
		ge.resetPlayerReadiness()
		ge.GameState.CurrentPhase = model.GamePhase_INCIDENTS
	}
}

func (ge *GameEngine) handleIncidentsPhase() {
	ge.logger.Info("Incidents Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)

	// Check for incidents on the current day
	for _, incident := range ge.gameConfig.GetIncidents() {
		if incident.Day == ge.GameState.CurrentDay {
			if ge.checkConditions(incident.TriggerConditions, nil, nil) {
				ge.logger.Info("Incident triggered!", zap.String("incident_name", incident.Name))
				incident := &model.Incident{Config: incident, Name: incident.Name, Day: incident.Day}
				ge.applyAndPublishEvent(model.GameEventType_INCIDENT_TRIGGERED, &model.IncidentTriggeredEvent{Incident: incident})
				// TODO: Apply incident effect
			}
		}
	}

	// This logic is a bit flawed, needs rework based on script structure
	// tragedyOccurred := false
	// for _, tragedy := range ge.GameState.Script.Tragedies {
	// 	// Check if the tragedy is active for the day and hasn't been prevented
	// 	if tragedy.Day == ge.GameState.CurrentDay && ge.GameState.ActiveTragedies[int32(tragedy.TragedyType)] && !ge.GameState.PreventedTragedies[int32(tragedy.TragedyType)] {
	// 		if ge.checkConditions(tragedy.Conditions) {
	// 			ge.logger.Info("Tragedy triggered!", zap.String("tragedy_type", string(tragedy.TragedyType)))
	// 			ge.publishGameEvent(model.GameEventType_TRAGEDY_TRIGGERED, &model.TragedyTriggeredEvent{TragedyType: tragedy.TragedyType})
	// 			tragedyOccurred = true
	// 			ge.GameState.ActiveTragedies[int32(tragedy.TragedyType)] = false // A tragedy can only occur once
	// 			break                                                            // Only one tragedy per day
	// 		}
	// 	}
	// }

	if gameOver, winner := ge.checkGameEndConditions(); gameOver {
		ge.endGame(winner)
		return
	}

	// if tragedyOccurred {
	// 	ge.GameState.CurrentPhase = model.GamePhase_LOOP_END
	// } else {
	ge.GameState.CurrentPhase = model.GamePhase_DAY_END
	// }
}

func (ge *GameEngine) handleDayEndPhase() {
	ge.logger.Info("Day End Phase", zap.Int("day", int(ge.GameState.CurrentDay)))
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

func (ge *GameEngine) handleLoopEndPhase() {
	ge.logger.Info("Loop End Phase", zap.Int("loop", int(ge.GameState.CurrentLoop)))
	ge.checkAndTriggerAbilities(model.TriggerType_ON_LOOP_START)

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

	// If it's the last loop, the outcome is final.
	if ge.GameState.CurrentLoop >= script.LoopCount {
		// If no one has won by the final loop, Protagonists win by default
		ge.endGame(model.PlayerRole_PROTAGONIST)
		return
	}

	// If no tragedy occurred and more loops are left, reset for the next loop.
	ge.resetLoop()
	ge.GameState.CurrentLoop++
	ge.GameState.CurrentDay = 1
	ge.GameState.CurrentPhase = model.GamePhase_SETUP
	ge.applyAndPublishEvent(model.GameEventType_LOOP_RESET, &model.LoopResetEvent{Loop: ge.GameState.CurrentLoop})
}

func (ge *GameEngine) handleProtagonistGuessPhase() {
	ge.logger.Info("Protagonist Guess Phase")
	// This phase is triggered by a player action (ActionMakeGuess).
	// The logic is handled in `handleMakeGuessAction`.
	// After the guess, the game transitions to GameOver.
	ge.GameState.CurrentPhase = model.GamePhase_GAME_OVER
}
