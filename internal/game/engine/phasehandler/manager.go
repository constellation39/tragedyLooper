package phasehandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// Manager orchestrates the game's lifecycle by executing phases determined by a flowchart.
type Manager struct {
	engine        GameEngine
	logger        *zap.Logger
	currentPhase  Phase
	timeoutTarget int64 // The tick count at which the current phase will time out.
	gameStarted   bool
	flowchart     *FlowchartManager
}

// NewManager creates a new phase manager.
func NewManager(engine GameEngine) *Manager {
	pm := &Manager{
		engine:       engine,
		logger:       engine.Logger().Named("Manager"),
		currentPhase: GetPhase(model.GamePhase_GAME_PHASE_SETUP),
		flowchart:    NewFlowchartManager(engine),
	}
	return pm
}

// Start begins the phase lifecycle by transitioning to the initial phase.
func (pm *Manager) Start() {
	pm.transitionTo(pm.currentPhase.Type())
}

// OnTick is called periodically by the game engine to check for phase timeouts.
func (pm *Manager) OnTick() {
	if pm.timeoutTarget > 0 && pm.engine.GetGameState().Tick >= pm.timeoutTarget {
		pm.HandleTimeout()
	}
}

// CurrentPhase returns the current game phase.
func (pm *Manager) CurrentPhase() Phase {
	return pm.currentPhase
}

// HandleAction delegates the action to the current phase and then attempts a transition if the phase is ready.
func (pm *Manager) HandleAction(player *model.Player, action *model.PlayerActionPayload) {
	if pm.currentPhase.HandleAction(pm.engine, player, action) {
		pm.transitionToNext()
	}
}

// HandleEvent delegates the event to the current phase and then attempts a transition if the phase is ready.
func (pm *Manager) HandleEvent(event *model.GameEvent) {
	if pm.currentPhase.HandleEvent(pm.engine, event) {
		pm.transitionToNext()
	}
}

// HandleTimeout handles a phase timeout and then attempts a transition.
func (pm *Manager) HandleTimeout() {
	pm.currentPhase.HandleTimeout(pm.engine)
	pm.transitionToNext()
}

// transitionTo handles the logic of moving from one game phase to another.
func (pm *Manager) transitionTo(nextPhaseType model.GamePhase) bool {
	if nextPhaseType == model.GamePhase_GAME_PHASE_UNSPECIFIED || (pm.gameStarted && nextPhaseType == pm.currentPhase.Type()) {
		return false
	}

	nextPhase := GetPhase(nextPhaseType)
	if nextPhase == nil {
		pm.logger.Error("Failed to get next phase", zap.String("phase", nextPhaseType.String()))
		return false
	}

	pm.timeoutTarget = 0

	if pm.gameStarted {
		pm.logger.Info("Transitioning phase", zap.String("from", pm.currentPhase.Type().String()), zap.String("to", nextPhase.Type().String()))
		pm.currentPhase.Exit(pm.engine)
	} else {
		pm.logger.Info("Entering initial phase", zap.String("to", nextPhase.Type().String()))
		pm.gameStarted = true
	}

	pm.currentPhase = nextPhase
	pm.engine.GetGameState().CurrentPhase = nextPhase.Type()

	// Enter the new phase.
	pm.currentPhase.Enter(pm.engine)

	// Set the timer for the new phase.
	ticks := pm.currentPhase.TimeoutTicks()
	if ticks > 0 {
		pm.timeoutTarget = pm.engine.GetGameState().Tick + ticks
	}

	// After entering, immediately check if we should transition again.
	// This handles auto-advancing phases.
	return pm.transitionToNext()
}

// transitionToNext determines the next phase from the flowchart and transitions to it.
func (pm *Manager) transitionToNext() bool {
	nextPhaseType := pm.flowchart.GetNextPhase(pm.currentPhase.Type())
	return pm.transitionTo(nextPhaseType)
}
