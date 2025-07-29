package engine // 定义游戏引擎包

import (
	"time"
	"tragedylooper/internal/game/engine/eventhandler"
	"tragedylooper/internal/game/engine/phase" // 导入阶段包
	model "tragedylooper/pkg/proto/v1"         // 导入协议缓冲区模型

	"go.uber.org/zap" // 导入 Zap 日志库
)

// phaseManager 负责管理游戏的阶段生命周期，包括转换和超时。
// 它封装了以前在 GameEngine 中的逻辑，从而实现了更清晰的关注点分离。
type phaseManager struct {
	engine       *GameEngine // 对父引擎的引用，以访问游戏状态和其他组件。
	logger       *zap.Logger // 日志记录器
	currentPhase phase.Phase // 当前的游戏阶段。
	phaseTimer   *time.Timer // 用于管理阶段超时的计时器。
	gameStarted  bool        // 标记游戏是否已开始。
}

// newPhaseManager 创建一个新的阶段管理器。
// engine: 游戏引擎的引用。
// 返回值: 新的 phaseManager 实例。
func newPhaseManager(engine *GameEngine) *phaseManager {
	pm := &phaseManager{
		engine:       engine,
		logger:       engine.logger.Named("PhaseManager"),
		currentPhase: &phase.SetupPhase{},
		phaseTimer:   time.NewTimer(time.Hour), // 用一个很长的时间初始化。
	}
	pm.phaseTimer.Stop() // 立即停止它；它将在第一次转换时重置。
	return pm
}

// start 开始阶段生命周期，转换到初始阶段。
func (pm *phaseManager) start() {
	pm.transitionTo(pm.currentPhase)
}

// stop 清理并停止阶段管理器的计时器。
func (pm *phaseManager) stop() {
	pm.phaseTimer.Stop()
}

// timer 返回阶段计时器的通道。
// 返回值: 计时器通道。
func (pm *phaseManager) timer() <-chan time.Time {
	return pm.phaseTimer.C
}

// CurrentPhase 返回当前的游戏阶段。
// 返回值: 当前阶段的接口。
func (pm *phaseManager) CurrentPhase() phase.Phase {
	return pm.currentPhase
}

// handleAction 将操作委托给当前阶段并转换到下一个阶段。
// playerID: 执行操作的玩家ID。
// action: 玩家操作的负载。
func (pm *phaseManager) handleAction(playerID int32, action *model.PlayerActionPayload) {
	nextPhase := pm.currentPhase.HandleAction(pm.engine, playerID, action)
	pm.transitionTo(nextPhase)
}

// handleEvent 将事件委托给当前阶段并转换到下一个阶段。
// event: 接收到的游戏事件。
func (pm *phaseManager) handleEvent(event *model.GameEvent) {
	// Step 1: Synchronously apply state changes.
	// This is crucial for ensuring the game state is consistent before any other logic runs.
	if err := eventhandler.ProcessEvent(pm.engine.GameState, event); err != nil {
		pm.logger.Error("Failed to apply event to game state", zap.Error(err), zap.String("type", event.Type.String()))
		// We continue even if the handler fails, to allow phases and listeners to react.
	}

	nextPhase := pm.currentPhase.HandleEvent(pm.engine, event)
	pm.transitionTo(nextPhase)
}

// handleTimeout 处理阶段超时并转换到下一个阶段。
func (pm *phaseManager) handleTimeout() {
	nextPhase := pm.currentPhase.HandleTimeout(pm.engine)
	pm.transitionTo(nextPhase)
}

// transitionTo 处理从一个游戏阶段移动到另一个游戏阶段的逻辑。
// 它使用一个循环来处理连续的即时阶段转换，而无需递归。
// nextPhase: 要转换到的下一个阶段的实例。
func (pm *phaseManager) transitionTo(nextPhase phase.Phase) {
	// nil 的 nextPhase 表示不需要状态更改。
	if nextPhase == nil {
		return
	}

	// 循环处理一连串的即时阶段转换（例如，设置 -> 主要 -> 行动）。
	// 这避免了如果一个阶段的 Enter() 方法立即返回一个新阶段而导致的深度递归。
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
		pm.engine.GameState.CurrentPhase = nextPhase.Type() // 引擎仍然拥有状态。

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
}
