package engine

import (
	"time"
	"tragedylooper/internal/game/engine/phase"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// phaseManager is responsible for managing the game's phase lifecycle, including transitions and timeouts.
// It encapsulates the logic that was previously in the GameEngine, leading to a cleaner separation of concerns.
type phaseManager struct {
	engine       *GameEngine // A reference to the parent engine to access game state and other components.
	logger       *zap.Logger
	currentPhase phase.Phase
	phaseTimer   *time.Timer
	gameStarted  bool
}

// newPhaseManager creates a new phase manager.
func newPhaseManager(engine *GameEngine) *phaseManager {
	pm := &phaseManager{
		engine:       engine,
		logger:       engine.logger.Named("PhaseManager"),
		currentPhase: &phase.SetupPhase{},
		phaseTimer:   time.NewTimer(time.Hour), // Initialized with a long duration.
	}
	pm.phaseTimer.Stop() // Stop it immediately; it will be reset on the first transition.
	return pm
}

// start begins the phase lifecycle, transitioning to the initial phase.
func (pm *phaseManager) start() {
	pm.transitionTo(pm.currentPhase)
}

// stop cleanly stops the phase manager's timer.
func (pm *phaseManager) stop() {
	pm.phaseTimer.Stop()
}

// timer returns the channel for the phase timer.
func (pm *phaseManager) timer() <-chan time.Time {
	return pm.phaseTimer.C
}

func (pm *phaseManager) CurrentPhase() phase.Phase {
	return pm.currentPhase
}

// handleAction delegates an action to the current phase and transitions to the next.
func (pm *phaseManager) handleAction(playerID int32, action *model.PlayerActionPayload) {
	nextPhase := pm.currentPhase.HandleAction(pm.engine, playerID, action)
	pm.transitionTo(nextPhase)
}

// handleEvent delegates an event to the current phase and transitions to the next.
func (pm *phaseManager) handleEvent(event *model.GameEvent) {
	nextPhase := pm.currentPhase.HandleEvent(pm.engine, event)
	pm.transitionTo(nextPhase)
}

// handleTimeout handles a phase timeout and transitions to the next phase.
func (pm *phaseManager) handleTimeout() {
	nextPhase := pm.currentPhase.HandleTimeout(pm.engine)
	pm.transitionTo(nextPhase)
}

// transitionTo handles the logic of moving from one game phase to another.
// It uses a loop to process immediate, consecutive phase transitions without recursion.
func (pm *phaseManager) transitionTo(nextPhase phase.Phase) {
	// A nil nextPhase indicates no state change is needed.
	if nextPhase == nil {
		return
	}

	// Loop to handle a chain of immediate phase transitions (e.g., Setup -> Main -> Action).
	// This avoids deep recursion if a phase's Enter() method immediately returns a new phase.
	for nextPhase != nil {
		pm.phaseTimer.Stop()

		if pm.gameStarted {
			pm.logger.Info("Transitioning phase", zap.String("from", pm.currentPhase.Type().String()), zap.String("to", nextPhase.Type().String()))
			pm.currentPhase.Exit(pm.engine)
		} else {
			pm.logger.Info("Entering initial phase", zap.String("to", nextPhase.Type().String()))
			pm.gameStarted = true
		}

		pm.currentPhase = nextPhase
		pm.engine.GameState.CurrentPhase = nextPhase.Type() // The engine still owns the state.

		// Enter the new phase. It might return another phase to transition to immediately.
		followingPhase := pm.currentPhase.Enter(pm.engine)

		// Set the timer for the new phase. If the duration is 0, the timer remains stopped.
		duration := pm.currentPhase.TimeoutDuration()
		if duration > 0 {
			pm.phaseTimer.Reset(duration)
		}

		// The loop continues with the next phase, if any.
		nextPhase = followingPhase
	}
}
