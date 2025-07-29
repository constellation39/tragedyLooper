package engine

import (
	"time"
	model "tragedylooper/internal/game/proto/v1"
)

// Phase is an interface for a game phase.
type Phase interface {
	Type() model.GamePhase
	// Enter is called when the game transitions into this phase.
	Enter(ge *GameEngine)
	// HandleEvent processes a game event. It returns the next phase to transition to, or nil to stay in the current phase.
	HandleEvent(ge *GameEngine, event *model.GameEvent) Phase
	// HandleTimeout is called when the phase timer expires. It returns the next phase to transition to, or nil.
	HandleTimeout(ge *GameEngine) Phase
	// Exit is called when the game transitions out of this phase.
	Exit(ge *GameEngine)
	// TimeoutDuration returns the duration after which the phase should time out.
	TimeoutDuration() time.Duration
}

var phaseImplementations = map[model.GamePhase]Phase{
	model.GamePhase_SETUP:              &SetupPhase{},
	model.GamePhase_MASTERMIND_SETUP:   &MastermindSetupPhase{},
	model.GamePhase_LOOP_START:         &LoopStartPhase{},
	model.GamePhase_DAY_START:          &DayStartPhase{},
	model.GamePhase_CARD_PLAY:          &CardPlayPhase{},
	model.GamePhase_CARD_REVEAL:        &CardRevealPhase{},
	model.GamePhase_CARD_RESOLVE:       &CardResolvePhase{},
	model.GamePhase_ABILITIES:          &AbilitiesPhase{},
	model.GamePhase_INCIDENTS:          &IncidentsPhase{},
	model.GamePhase_DAY_END:            &DayEndPhase{},
	model.GamePhase_LOOP_END:           &LoopEndPhase{},
	model.GamePhase_PROTAGONIST_GUESS:  &ProtagonistGuessPhase{},
	model.GamePhase_GAME_OVER:          &GameOverPhase{},
}

// --- Base Phase --- (for embedding)
type basePhase struct{}

func (p *basePhase) Enter(ge *GameEngine)                 {}
func (p *basePhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase { return nil }
func (p *basePhase) HandleTimeout(ge *GameEngine) Phase   { return nil }
func (p *basePhase) Exit(ge *GameEngine)                  {}
func (p *basePhase) TimeoutDuration() time.Duration       { return 0 } // No timeout by default

// --- SetupPhase ---
type SetupPhase struct{ basePhase }

func (p *SetupPhase) Type() model.GamePhase { return model.GamePhase_SETUP }
func (p *SetupPhase) Enter(ge *GameEngine) {
	// TODO: Implement logic for Mastermind to choose sub-scenario and place characters.
	// For now, we transition directly.
	ge.transitionTo(phaseImplementations[model.GamePhase_MASTERMIND_SETUP])
}

// --- MastermindSetupPhase ---
type MastermindSetupPhase struct{ basePhase }

func (p *MastermindSetupPhase) Type() model.GamePhase { return model.GamePhase_MASTERMIND_SETUP }
func (p *MastermindSetupPhase) Enter(ge *GameEngine) {
	// TODO: Mastermind places characters and sets up their board.
	// For now, we transition directly.
	ge.transitionTo(phaseImplementations[model.GamePhase_LOOP_START])
}

// --- LoopStartPhase ---
type LoopStartPhase struct{ basePhase }

func (p *LoopStartPhase) Type() model.GamePhase { return model.GamePhase_LOOP_START }
func (p *LoopStartPhase) Enter(ge *GameEngine) {
	ge.GameState.CurrentLoop++
	ge.GameState.CurrentDay = 0
	ge.applyAndPublishEvent(model.GameEventType_LOOP_RESET, &model.LoopResetEvent{LoopNumber: ge.GameState.CurrentLoop})
	ge.transitionTo(phaseImplementations[model.GamePhase_DAY_START])
}

// --- DayStartPhase ---
type DayStartPhase struct{ basePhase }

func (p *DayStartPhase) Type() model.GamePhase { return model.GamePhase_DAY_START }
func (p *DayStartPhase) Enter(ge *GameEngine) {
	ge.GameState.CurrentDay++
	ge.GameState.PlayedCardsThisDay = make(map[int32]*model.Card)
	ge.resetPlayerReadiness()
	ge.applyAndPublishEvent(model.GameEventType_DAY_ADVANCED, &model.DayAdvancedEvent{Day: ge.GameState.CurrentDay, Loop: ge.GameState.CurrentLoop})
	ge.transitionTo(phaseImplementations[model.GamePhase_CARD_PLAY])
}

// --- CardPlayPhase ---
type CardPlayPhase struct{ basePhase }

func (p *CardPlayPhase) Type() model.GamePhase { return model.GamePhase_CARD_PLAY }
func (p *CardPlayPhase) Enter(ge *GameEngine) {
	// Players have a certain amount of time to play their cards.
}
func (p *CardPlayPhase) HandleTimeout(ge *GameEngine) Phase {
	// If players don't act in time, we might auto-pass for them.
	return phaseImplementations[model.GamePhase_CARD_REVEAL]
}
func (p *CardPlayPhase) HandleEvent(ge *GameEngine, event *model.GameEvent) Phase {
	if ge.areAllPlayersReady() {
		return phaseImplementations[model.GamePhase_CARD_REVEAL]
	}
	return nil
}
func (p *CardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second } // Example timeout

// --- CardRevealPhase ---
type CardRevealPhase struct{ basePhase }

func (p *CardRevealPhase) Type() model.GamePhase { return model.GamePhase_CARD_REVEAL }
func (p *CardRevealPhase) Enter(ge *GameEngine) {
	// Reveal all cards played this turn.
	ge.applyAndPublishEvent(model.GameEventType_CARD_REVEALED, &model.CardRevealedEvent{Cards: ge.GameState.PlayedCardsThisDay})
	ge.transitionTo(phaseImplementations[model.GamePhase_CARD_RESOLVE])
}

// --- CardResolvePhase ---
type CardResolvePhase struct{ basePhase }

func (p *CardResolvePhase) Type() model.GamePhase { return model.GamePhase_CARD_RESOLVE }
func (p *CardResolvePhase) Enter(ge *GameEngine) {
	ge.resolveMovement()
	ge.resolveOtherCards()
	ge.transitionTo(phaseImplementations[model.GamePhase_ABILITIES])
}

// --- AbilitiesPhase ---
type AbilitiesPhase struct{ basePhase }

func (p *AbilitiesPhase) Type() model.GamePhase { return model.GamePhase_ABILITIES }
func (p *AbilitiesPhase) Enter(ge *GameEngine) {
	// Players can use abilities.
	// This phase might require player input and have a timeout.
	ge.transitionTo(phaseImplementations[model.GamePhase_INCIDENTS])
}

// --- IncidentsPhase ---
type IncidentsPhase struct{ basePhase }

func (p *IncidentsPhase) Type() model.GamePhase { return model.GamePhase_INCIDENTS }
func (p *IncidentsPhase) Enter(ge *GameEngine) {
	ge.triggerIncidents()
	ge.transitionTo(phaseImplementations[model.GamePhase_DAY_END])
}

// --- DayEndPhase ---
type DayEndPhase struct{ basePhase }

func (p *DayEndPhase) Type() model.GamePhase { return model.GamePhase_DAY_END }
func (p *DayEndPhase) Enter(ge *GameEngine) {
	if ge.GameState.CurrentDay >= ge.GameState.DaysPerLoop {
		ge.transitionTo(phaseImplementations[model.GamePhase_LOOP_END])
	} else {
		ge.transitionTo(phaseImplementations[model.GamePhase_DAY_START])
	}
}

// --- LoopEndPhase ---
type LoopEndPhase struct{ basePhase }

func (p *LoopEndPhase) Type() model.GamePhase { return model.GamePhase_LOOP_END }
func (p *LoopEndPhase) Enter(ge *GameEngine) {
	if ge.GameState.CurrentLoop >= ge.gameConfig.GetScript().LoopCount {
		// Protagonists get a final chance to guess
		ge.transitionTo(phaseImplementations[model.GamePhase_PROTAGONIST_GUESS])
	} else {
		ge.transitionTo(phaseImplementations[model.GamePhase_LOOP_START])
	}
}

// --- ProtagonistGuessPhase ---
type ProtagonistGuessPhase struct{ basePhase }

func (p *ProtagonistGuessPhase) Type() model.GamePhase { return model.GamePhase_PROTAGONIST_GUESS }

// --- GameOverPhase ---
type GameOverPhase struct{ basePhase }

func (p *GameOverPhase) Type() model.GamePhase { return model.GamePhase_GAME_OVER }
func (p *GameOverPhase) Enter(ge *GameEngine) {
	// Clean up, announce winner, etc.
	ge.StopGameLoop()
}