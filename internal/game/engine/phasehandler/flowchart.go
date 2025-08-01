package phasehandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// Transition defines a potential move from one phase to another.
type Transition struct {
	// Next is the phase to transition to.
	Next model.GamePhase
	// Condition is a function that must return true for the transition to be taken.
	// If Condition is nil, it is considered a default/unconditional path.
	Condition func(ge GameEngine) bool
}

// Flowchart defines the entire game flow as a state machine.
// For each phase, it lists possible transitions. The first transition whose condition is met will be taken.
var Flowchart = map[model.GamePhase][]Transition{
	model.GamePhase_GAME_PHASE_SETUP: {
		{Next: model.GamePhase_GAME_PHASE_MASTERMIND_SETUP},
	},
	model.GamePhase_GAME_PHASE_MASTERMIND_SETUP: {
		{Next: model.GamePhase_GAME_PHASE_LOOP_START},
	},
	model.GamePhase_GAME_PHASE_LOOP_START: {
		{Next: model.GamePhase_GAME_PHASE_DAY_START},
	},
	model.GamePhase_GAME_PHASE_DAY_START: {
		{Next: model.GamePhase_GAME_PHASE_MASTERMIND_CARD_PLAY},
	},
	model.GamePhase_GAME_PHASE_MASTERMIND_CARD_PLAY: {
		{Next: model.GamePhase_GAME_PHASE_PROTAGONIST_CARD_PLAY},
	},
	model.GamePhase_GAME_PHASE_PROTAGONIST_CARD_PLAY: {
		{Next: model.GamePhase_GAME_PHASE_CARD_REVEAL},
	},
	model.GamePhase_GAME_PHASE_CARD_REVEAL: {
		{Next: model.GamePhase_GAME_PHASE_CARD_RESOLVE},
	},
	model.GamePhase_GAME_PHASE_CARD_RESOLVE: {
		{Next: model.GamePhase_GAME_PHASE_ABILITIES},
	},
	model.GamePhase_GAME_PHASE_ABILITIES: {
		{Next: model.GamePhase_GAME_PHASE_INCIDENTS},
	},
	model.GamePhase_GAME_PHASE_INCIDENTS: {
		{Next: model.GamePhase_GAME_PHASE_DAY_END},
	},
	model.GamePhase_GAME_PHASE_DAY_END: {
		{
			Next: model.GamePhase_GAME_PHASE_LOOP_END,
			Condition: func(ge GameEngine) bool {
				// Placeholder for loop failure condition
				// For example: return ge.GetGameState().IsLoopOver
				return false
			},
		},
		{
			Next: model.GamePhase_GAME_PHASE_DAY_START, // Next day
		},
	},
	model.GamePhase_GAME_PHASE_LOOP_END: {
		{
			Next: model.GamePhase_GAME_PHASE_GAME_OVER,
			Condition: func(ge GameEngine) bool {
				// Placeholder for game over condition
				// For example: return ge.GetGameState().CurrentLoop >= ge.GetGameRepo().Script().MaxLoops
				return false
			},
		},
		{
			Next: model.GamePhase_GAME_PHASE_LOOP_START, // Next loop
		},
	},
	// PROTAGONIST_GUESS and GAME_OVER are terminal or special phases, handled separately.
}

// FlowchartManager uses the Flowchart to determine phase transitions.
type FlowchartManager struct {
	engine GameEngine
}

// NewFlowchartManager creates a new manager for the game's flow.
func NewFlowchartManager(engine GameEngine) *FlowchartManager {
	return &FlowchartManager{engine: engine}
}

// GetNextPhase finds the next logical phase based on the flowchart and current game state.
func (fm *FlowchartManager) GetNextPhase(currentPhase model.GamePhase) model.GamePhase {
	transitions, ok := Flowchart[currentPhase]
	if !ok {
		return model.GamePhase_GAME_PHASE_UNSPECIFIED
	}

	for _, transition := range transitions {
		if transition.Condition == nil || transition.Condition(fm.engine) {
			return transition.Next
		}
	}

	return model.GamePhase_GAME_PHASE_UNSPECIFIED
}
