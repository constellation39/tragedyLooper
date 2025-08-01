package phasehandler

import (
	"time"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// phaseManager 负责管理游戏的阶段生命周期，包括转换和超时。
// 它封装了以前在 GameEngine 中的逻辑，从而实现了更清晰的关注点分离。
type Manager struct {
	engine       GameEngine  // 对父引擎的引用，以访问游戏状态和其他组件。
	logger       *zap.Logger // 日志记录器
	currentPhase Phase       // 当前的游戏阶段。
	phaseTimer   *time.Timer // 用于管理阶段超时的计时器。
	gameStarted  bool        // 标记游戏是否已开始。
}

// newPhaseManager 创建一个新的阶段管理器。
func NewManager(engine GameEngine) *Manager {
	pm := &Manager{
		engine:       engine,
		logger:       engine.Logger().Named("Manager"),
		currentPhase: GetPhase(model.GamePhase_GAME_PHASE_SETUP), // 初始阶段从注册表获取
		phaseTimer:   time.NewTimer(time.Hour),                   // 用一个很长的时间初始化。
	}
	pm.phaseTimer.Stop() // 立即停止它；它将在第一次转换时重置。
	return pm
}

// start 开始阶段生命周期，转换到初始阶段。
func (pm *Manager) Start() {
	pm.transitionTo(pm.currentPhase)
}

// OnTick 由游戏引擎定期调用。
// 它会检查阶段超时并触发相应的处理程序。
// 这将超时检查逻辑从引擎的主 select 循环移入管理器，从而封装了与阶段相关的计时。
func (pm *Manager) OnTick() bool {
	select {
	case <-pm.phaseTimer.C:
		return pm.HandleTimeout()
	default:
		// 非阻塞：计时器尚未触发。
		return false
	}
}

// CurrentPhase 返回当前的游戏阶段。
func (pm *Manager) CurrentPhase() Phase {
	return pm.currentPhase
}

// HandleAction 将操作委托给当前阶段并转换到下一个阶段。
func (pm *Manager) HandleAction(player *model.Player, action *model.PlayerActionPayload) bool {
	nextPhase := pm.currentPhase.HandleAction(pm.engine, player, action)
	return pm.transitionTo(nextPhase)
}

// handleEvent 将事件委托给当前阶段并转换到下一个阶段。
func (pm *Manager) HandleEvent(event *model.GameEvent) bool {
	nextPhase := pm.currentPhase.HandleEvent(pm.engine, event)
	return pm.transitionTo(nextPhase)
}

// handleTimeout 处理阶段超时并转换到下一个阶段。
func (pm *Manager) HandleTimeout() bool {
	nextPhase := pm.currentPhase.HandleTimeout(pm.engine)
	return pm.transitionTo(nextPhase)
}

// transitionTo 处理从一个游戏阶段移动到另一个游戏阶段的逻辑。
// 它使用一个循环来处理连续的即时阶段转换，而无需递归。
func (pm *Manager) transitionTo(nextPhase Phase) bool {
	// nil 的 nextPhase 表示不需要状态更改。
	if nextPhase == nil {
		return false
	}

	// 循环处理一连串的即时阶段转换（例如，设置 -> 主要 -> 行动）。
	// 这避免了如果一个阶段的 Enter() 方法立即返回一个新阶段而导致的深度递归。
	for nextPhase != nil {
		pm.phaseTimer.Stop()

		if pm.gameStarted {
			pm.logger.Info("Transitioning phasehandler", zap.String("from", pm.currentPhase.Type().String()), zap.String("to", nextPhase.Type().String()))
			pm.currentPhase.Exit(pm.engine)
		} else {
			pm.logger.Info("Entering initial phasehandler", zap.String("to", nextPhase.Type().String()))
			pm.gameStarted = true
		}

		pm.currentPhase = nextPhase
		pm.engine.GetGameState().CurrentPhase = nextPhase.Type() // 引擎仍然拥有状态。

		// 进入新阶段。它可能会返回另一个要立即转换到的阶段。
		followingPhase := pm.currentPhase.Enter(pm.engine)

		// 为新阶段设置计时器。如果持续时间为 0，则计时器保持停止状态。
		duration := pm.currentPhase.TimeoutDuration()
		if duration > 0 {
			pm.phaseTimer.Reset(duration)
		}

		// 循环继续到下一个阶段（如果有）。
		nextPhase = followingPhase
	}
	return true
}
