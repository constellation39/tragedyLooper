package phase

import (
	"time"
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// GameEngine 定义了阶段可以与之交互的游戏引擎的接口。
type GameEngine interface {
	// TriggerEvent 应用事件并发布。
	TriggerEvent(eventType model.GameEventType, eventData *model.EventPayload)
	// AreAllPlayersReady 检查所有玩家是否都已准备好。
	AreAllPlayersReady() bool
	// CheckCondition 检查条件是否满足。
	CheckCondition(condition *model.Condition) (bool, error)
	// Logger 返回游戏引擎的日志记录器。
	Logger() *zap.Logger
	// ResetPlayerReadiness 重置所有玩家的准备状态。
	ResetPlayerReadiness()
	// SetPlayerReady 设置指定玩家的准备状态为 true。
	SetPlayerReady(playerID int32)
	// StopGameLoop 停止游戏主循环。
	StopGameLoop()
	// TriggerIncidents 触发事件。
	TriggerIncidents()
	// GetGameState 返回当前游戏状态。
	GetGameState() *model.GameState
	// GetGameRepo 返回游戏配置。
	GetGameRepo() loader.GameConfig
	// GetCharacterByID 根据角色ID获取角色对象。
	GetCharacterByID(id int32) *model.Character
	// MoveCharacter 移动角色。
	MoveCharacter(char *model.Character, dx, dy int)
	GetMastermindPlayer() *model.Player
	GetProtagonistPlayers() []*model.Player
	ApplyEffect(effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error
}

// Phase 是游戏阶段的接口，定义了每个阶段必须实现的方法。
type Phase interface {
	// Type 返回阶段的类型。
	Type() model.GamePhase
	// Enter 在进入此阶段时调用，返回下一个阶段（如果立即切换）。
	Enter(ge GameEngine) Phase
	// HandleAction 处理玩家在此阶段的操作，返回下一个阶段（如果发生变化）。
	HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase
	// HandleEvent 处理在此阶段接收到的游戏事件，返回下一个阶段（如果发生变化）。
	HandleEvent(ge GameEngine, event *model.GameEvent) Phase
	// HandleTimeout 处理此阶段的超时，返回下一个阶段。
	HandleTimeout(ge GameEngine) Phase
	// Exit 在退出此阶段时调用。
	Exit(ge GameEngine)
	// TimeoutDuration 返回此阶段的超时持续时间。
	TimeoutDuration() time.Duration
}

// basePhase 是一个辅助结构体，为 Phase 接口提供默认实现，方便嵌入和重写。
type basePhase struct{}

// Enter 是 Phase 接口的默认实现，不执行任何操作并返回 nil。
func (p *basePhase) Enter(ge GameEngine) Phase { return nil }

// HandleAction 是 Phase 接口的默认实现，不执行任何操作并返回 nil。
func (p *basePhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	return nil
}

// HandleEvent 是 Phase 接口的默认实现，不执行任何操作并返回 nil。
func (p *basePhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }

// HandleTimeout 是 Phase 接口的默认实现，不执行任何操作并返回 nil。
func (p *basePhase) HandleTimeout(ge GameEngine) Phase { return nil }

// Exit 是 Phase 接口的默认实现，不执行任何操作。
func (p *basePhase) Exit(ge GameEngine) {}

// TimeoutDuration 是 Phase 接口的默认实现，返回 0，表示没有超时。
func (p *basePhase) TimeoutDuration() time.Duration { return 0 }
