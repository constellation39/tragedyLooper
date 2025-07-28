package engine

import (
	"time"
	model "tragedylooper/internal/game/proto/v1"

	"go.uber.org/zap"
)

const defaultPhaseTimeout = 30 * time.Second

// Phase represents a distinct phase in the game loop.
type Phase interface {
	// Enter is called when the game enters this phase.
	Enter(ge *GameEngine)
	// HandleEvent processes a game event. It returns the next phase to transition to,
	// or nil to remain in the current phase.
	HandleEvent(ge *GameEngine, event *model.GameEvent) Phase
	// HandleTimeout is called when the phase times out.
	HandleTimeout(ge *GameEngine) Phase
	// Exit is called when the game leaves this phase.
	Exit(ge *GameEngine)
	// Type returns the enum type of the phase.
	Type() model.GamePhase
	// TimeoutDuration returns the duration after which the phase should time out.
	TimeoutDuration() time.Duration
}

// phaseImplementations is a map of all phase types to their implementations.
var phaseImplementations = map[model.GamePhase]Phase{
	model.GamePhase_SETUP:             &SetupPhase{},
	model.GamePhase_MASTERMIND_SETUP:  &MastermindSetupPhase{},
	model.GamePhase_CARD_PLAY:         &CardPlayPhase{},
	model.GamePhase_CARD_REVEAL:       &CardRevealPhase{},
	model.GamePhase_CARD_RESOLVE:      &CardResolvePhase{},
	model.GamePhase_ABILITIES:         &AbilitiesPhase{},
	model.GamePhase_INCIDENTS:         &IncidentsPhase{},
	model.GamePhase_DAY_END:           &DayEndPhase{},
	model.GamePhase_LOOP_END:          &LoopEndPhase{},
	model.GamePhase_PROTAGONIST_GUESS: &ProtagonistGuessPhase{},
	model.GamePhase_GAME_OVER:         &GameOverPhase{},
}

// --- Setup Phase ---

type SetupPhase struct{}

func (p *SetupPhase) Enter(ge *GameEngine) {
	ge.logger.Info("Entering Setup Phase")
	ge.resetPlayerReadiness()
	ge.GameState.PlayedCardsThisDay = make(map[int32]*model.Card) // Clear cards for the new day
	ge.checkAndTriggerAbilities(model.TriggerType_ON_DAY_START)
	ge.applyAndPublishEvent(model.GameEventType_DAY_ADVANCED, &model.DayAdvancedEvent{Day: ge.GameState.CurrentDay, Loop: ge.GameState.CurrentLoop})
}

func (p *SetupPhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	return phaseImplementations[model.GamePhase_CARD_PLAY]
}

func (p *SetupPhase) HandleTimeout(ge *GameEngine) Phase {
	return nil // No timeout for this phase
}

func (p *SetupPhase) Exit(ge *GameEngine) {
	ge.logger.Info("Exiting Setup Phase")
}

func (p *SetupPhase) Type() model.GamePhase {
	return model.GamePhase_SETUP
}

func (p *SetupPhase) TimeoutDuration() time.Duration {
	return 0 // No timeout
}

// --- Mastermind Setup Phase ---

type MastermindSetupPhase struct{}

func (p *MastermindSetupPhase) Enter(ge *GameEngine) {
	ge.logger.Info("Entering Mastermind Setup Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)
	mastermind := ge.getMastermindPlayer()
	if mastermind != nil && mastermind.IsLlm {
		go ge.triggerLLMPlayerAction(mastermind.Id)
	}
}

func (p *MastermindSetupPhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	if event.Type == model.GameEventType_PLAYER_ACTION && ge.isMastermindReady() {
		return phaseImplementations[model.GamePhase_SETUP]
	}
	return nil // Remain in this phase
}

func (p *MastermindSetupPhase) HandleTimeout(ge *GameEngine) Phase {
	ge.logger.Warn("Mastermind setup phase timed out. Proceeding automatically.")
	return phaseImplementations[model.GamePhase_SETUP]
}

func (p *MastermindSetupPhase) Exit(ge *GameEngine) {
	ge.logger.Info("Exiting Mastermind Setup Phase")
	ge.resetPlayerReadiness()
}
func (p *MastermindSetupPhase) Type() model.GamePhase {
	return model.GamePhase_MASTERMIND_SETUP
}

func (p *MastermindSetupPhase) TimeoutDuration() time.Duration {
	return defaultPhaseTimeout
}

// --- Card Play Phase ---

type CardPlayPhase struct{}

func (p *CardPlayPhase) Enter(ge *GameEngine) {
	ge.logger.Info("Entering Card Play Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)
	for playerID, player := range ge.GameState.Players {
		if !ge.playerReady[playerID] && player.IsLlm {
			go ge.triggerLLMPlayerAction(playerID)
		}
	}
}

func (p *CardPlayPhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	if event.Type == model.GameEventType_PLAYER_ACTION && ge.areAllPlayersReady() {
		return phaseImplementations[model.GamePhase_CARD_REVEAL]
	}
	return nil // Remain in this phase
}

func (p *CardPlayPhase) HandleTimeout(ge *GameEngine) Phase {
	ge.logger.Warn("Card play phase timed out. Proceeding with played cards.")
	return phaseImplementations[model.GamePhase_CARD_REVEAL]
}

func (p *CardPlayPhase) Exit(ge *GameEngine) {
	ge.logger.Info("Exiting Card Play Phase")
	ge.resetPlayerReadiness()
}
func (p *CardPlayPhase) Type() model.GamePhase {
	return model.GamePhase_CARD_PLAY
}

func (p *CardPlayPhase) TimeoutDuration() time.Duration {
	return defaultPhaseTimeout
}

// --- Card Reveal Phase ---

type CardRevealPhase struct{}

func (p *CardRevealPhase) Enter(ge *GameEngine) {
	ge.logger.Info("Entering Card Reveal Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)

	playedCards := make([]*model.Card, 0, len(ge.GameState.PlayedCardsThisDay))
	for _, card := range ge.GameState.PlayedCardsThisDay {
		playedCards = append(playedCards, card)
	}
	ge.applyAndPublishEvent(model.GameEventType_CARD_REVEALED, &model.CardRevealedEvent{Cards: playedCards})
}

func (p *CardRevealPhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	return phaseImplementations[model.GamePhase_CARD_RESOLVE]
}

func (p *CardRevealPhase) HandleTimeout(ge *GameEngine) Phase {
	return nil // No timeout
}

func (p *CardRevealPhase) Exit(ge *GameEngine) {
	ge.logger.Info("Exiting Card Reveal Phase")
}
func (p *CardRevealPhase) Type() model.GamePhase {
	return model.GamePhase_CARD_REVEAL
}

func (p *CardRevealPhase) TimeoutDuration() time.Duration {
	return 0 // No timeout
}

// --- Card Resolve Phase ---

type CardResolvePhase struct{}

func (p *CardResolvePhase) Enter(ge *GameEngine) {
	ge.logger.Info("Entering Card Resolve Phase")
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
}

func (p *CardResolvePhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	return phaseImplementations[model.GamePhase_ABILITIES]
}

func (p *CardResolvePhase) HandleTimeout(ge *GameEngine) Phase {
	return nil // No timeout
}

func (p *CardResolvePhase) Exit(ge *GameEngine) {
	ge.logger.Info("Exiting Card Resolve Phase")
}
func (p *CardResolvePhase) Type() model.GamePhase {
	return model.GamePhase_CARD_RESOLVE
}

func (p *CardResolvePhase) TimeoutDuration() time.Duration {
	return 0 // No timeout
}

// --- Abilities Phase ---

type AbilitiesPhase struct{}

func (p *AbilitiesPhase) Enter(ge *GameEngine) {
	ge.logger.Info("Entering Abilities Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)
	for playerID, player := range ge.GameState.Players {
		if !ge.playerReady[playerID] && player.IsLlm {
			go ge.triggerLLMPlayerAction(playerID)
		}
	}
}

func (p *AbilitiesPhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	if event.Type == model.GameEventType_PLAYER_ACTION && ge.areAllPlayersReady() {
		return phaseImplementations[model.GamePhase_INCIDENTS]
	}
	return nil // Remain in this phase
}

func (p *AbilitiesPhase) HandleTimeout(ge *GameEngine) Phase {
	ge.logger.Warn("Abilities phase timed out. Proceeding.")
	return phaseImplementations[model.GamePhase_INCIDENTS]
}

func (p *AbilitiesPhase) Exit(ge *GameEngine) {
	ge.logger.Info("Exiting Abilities Phase")
	ge.resetPlayerReadiness()
}
func (p *AbilitiesPhase) Type() model.GamePhase {
	return model.GamePhase_ABILITIES
}

func (p *AbilitiesPhase) TimeoutDuration() time.Duration {
	return defaultPhaseTimeout
}

// --- Incidents Phase ---

type IncidentsPhase struct{}

func (p *IncidentsPhase) Enter(ge *GameEngine) {
	ge.logger.Info("Entering Incidents Phase")
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
}

func (p *IncidentsPhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	if gameOver, winner := ge.checkGameEndConditions(); gameOver {
		ge.endGame(winner)
		return phaseImplementations[model.GamePhase_GAME_OVER]
	}
	return phaseImplementations[model.GamePhase_DAY_END]
}

func (p *IncidentsPhase) HandleTimeout(ge *GameEngine) Phase {
	return nil // No timeout
}

func (p *IncidentsPhase) Exit(ge *GameEngine) {
	ge.logger.Info("Exiting Incidents Phase")
}
func (p *IncidentsPhase) Type() model.GamePhase {
	return model.GamePhase_INCIDENTS
}

func (p *IncidentsPhase) TimeoutDuration() time.Duration {
	return 0 // No timeout
}

// --- Day End Phase ---

type DayEndPhase struct{}

func (p *DayEndPhase) Enter(ge *GameEngine) {
	ge.logger.Info("Entering Day End Phase", zap.Int("day", int(ge.GameState.CurrentDay)))
	ge.checkAndTriggerAbilities(model.TriggerType_ON_DAY_END)
	ge.GameState.CurrentDay++
}

func (p *DayEndPhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	script, err := ge.gameConfig.GetScript()
	if err != nil {
		ge.logger.Error("Failed to get script for day end phase", zap.Error(err))
		ge.endGame(model.PlayerRole_PLAYER_ROLE_UNSPECIFIED) // End game on error
		return phaseImplementations[model.GamePhase_GAME_OVER]
	}

	if ge.GameState.CurrentDay > script.DaysPerLoop {
		return phaseImplementations[model.GamePhase_LOOP_END]
	}
	return phaseImplementations[model.GamePhase_SETUP]
}

func (p *DayEndPhase) HandleTimeout(ge *GameEngine) Phase {
	return nil // No timeout
}

func (p *DayEndPhase) Exit(ge *GameEngine) {
	ge.logger.Info("Exiting Day End Phase")
}
func (p *DayEndPhase) Type() model.GamePhase {
	return model.GamePhase_DAY_END
}

func (p *DayEndPhase) TimeoutDuration() time.Duration {
	return 0 // No timeout
}

// --- Loop End Phase ---

type LoopEndPhase struct{}

func (p *LoopEndPhase) Enter(ge *GameEngine) {
	ge.logger.Info("Entering Loop End Phase", zap.Int("loop", int(ge.GameState.CurrentLoop)))
	ge.checkAndTriggerAbilities(model.TriggerType_ON_LOOP_END)
}

func (p *LoopEndPhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	if gameOver, winner := ge.checkGameEndConditions(); gameOver {
		ge.endGame(winner)
		return phaseImplementations[model.GamePhase_GAME_OVER]
	}

	script, err := ge.gameConfig.GetScript()
	if err != nil {
		ge.logger.Error("Failed to get script for loop end phase", zap.Error(err))
		ge.endGame(model.PlayerRole_PLAYER_ROLE_UNSPECIFIED) // End game on error
		return phaseImplementations[model.GamePhase_GAME_OVER]
	}

	if ge.GameState.CurrentLoop >= script.LoopCount {
		ge.logger.Info("Final loop has ended.")
		ge.applyAndPublishEvent(model.GameEventType_LOOP_WIN, &model.LoopWinEvent{Loop: ge.GameState.CurrentLoop})
		ge.endGame(model.PlayerRole_PROTAGONIST)
		return phaseImplementations[model.GamePhase_GAME_OVER]
	}

	ge.applyAndPublishEvent(model.GameEventType_LOOP_LOSS, &model.LoopLossEvent{Loop: ge.GameState.CurrentLoop})
	ge.resetLoop()
	ge.GameState.CurrentLoop++
	ge.GameState.CurrentDay = 1
	ge.applyAndPublishEvent(model.GameEventType_LOOP_RESET, &model.LoopResetEvent{Loop: ge.GameState.CurrentLoop})
	ge.checkAndTriggerAbilities(model.TriggerType_ON_LOOP_START)
	return phaseImplementations[model.GamePhase_MASTERMIND_SETUP]
}

func (p *LoopEndPhase) HandleTimeout(ge *GameEngine) Phase {
	return nil // No timeout
}

func (p *LoopEndPhase) Exit(ge *GameEngine) {
	ge.logger.Info("Exiting Loop End Phase")
}
func (p *LoopEndPhase) Type() model.GamePhase {
	return model.GamePhase_LOOP_END
}

func (p *LoopEndPhase) TimeoutDuration() time.Duration {
	return 0 // No timeout
}

// --- Protagonist Guess Phase ---

type ProtagonistGuessPhase struct{}

func (p *ProtagonistGuessPhase) Enter(ge *GameEngine) {
	ge.logger.Info("Entering Protagonist Guess Phase")
	ge.checkAndTriggerAbilities(model.TriggerType_ON_PHASE_START)
	for playerID, player := range ge.GameState.Players {
		if player.Role == model.PlayerRole_PROTAGONIST && !ge.playerReady[playerID] && player.IsLlm {
			go ge.triggerLLMPlayerAction(playerID)
		}
	}
}

func (p *ProtagonistGuessPhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	// The logic for handling the guess is in `actions.go:handleMakeGuessAction`.
	// That action will call `endGame` which transitions to the GameOverPhase.
	// We just wait here.
	return nil
}

func (p *ProtagonistGuessPhase) HandleTimeout(ge *GameEngine) Phase {
	ge.logger.Warn("Protagonist guess phase timed out. Mastermind wins.")
	ge.endGame(model.PlayerRole_MASTERMIND)
	return phaseImplementations[model.GamePhase_GAME_OVER]
}

func (p *ProtagonistGuessPhase) Exit(ge *GameEngine) {
	ge.logger.Info("Exiting Protagonist Guess Phase")
}
func (p *ProtagonistGuessPhase) Type() model.GamePhase {
	return model.GamePhase_PROTAGONIST_GUESS
}

func (p *ProtagonistGuessPhase) TimeoutDuration() time.Duration {
	return defaultPhaseTimeout
}

// --- Game Over Phase ---

type GameOverPhase struct{}

func (p *GameOverPhase) Enter(ge *GameEngine) {
	ge.logger.Info("Entering Game Over Phase. The game has concluded.")
}

func (p *GameOverPhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	return nil // No transitions out of game over
}

func (p *GameOverPhase) HandleTimeout(ge *GameEngine) Phase {
	return nil // No timeout
}

func (p *GameOverPhase) Exit(ge *GameEngine) {
	// This phase should not be exited.
}
func (p *GameOverPhase) Type() model.GamePhase {
	return model.GamePhase_GAME_OVER
}

func (p *GameOverPhase) TimeoutDuration() time.Duration {
	return 0 // No timeout
}
