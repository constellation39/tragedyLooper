package engine // 定义游戏引擎包

import (
	"tragedylooper/internal/game/engine/eventhandler" // 导入事件处理程序包
	model "tragedylooper/pkg/proto/v1" // 导入协议缓冲区模型

	"go.uber.org/zap" // 导入 Zap 日志库
	"google.golang.org/protobuf/proto" // 导入 Protobuf 核心库
	"google.golang.org/protobuf/types/known/anypb" // 导入 Any 类型，用于封装任意 Protobuf 消息
	"google.golang.org/protobuf/types/known/timestamppb" // 导入 Timestamp 类型
)

// eventManager 负责创建、处理和分派所有游戏事件。
// 它将事件生命周期与主 GameEngine 分离，确保了清晰且可维护的流程。
type eventManager struct {
	engine *GameEngine // 对父引擎的引用，以访问全局状态和管理器。
	logger *zap.Logger // 日志记录器

	// dispatchGameEvent 是一个出站通道，用于向外部侦听器广播已处理的游戏事件。
	dispatchGameEvent chan *model.GameEvent
}

// newEventManager 创建并返回一个新的 eventManager 实例。
// engine: 游戏引擎的引用。
func newEventManager(engine *GameEngine) *eventManager {
	return &eventManager{
		engine:            engine,
		logger:            engine.logger.Named("EventManager"),
		dispatchGameEvent: make(chan *model.GameEvent, 100),
	}
}

// createAndProcess 是创建、应用和广播游戏事件的中心方法。
// 它确保了一致的操作顺序：
// 1. 事件是根据有效负载创建的。
// 2. 游戏状态由事件处理程序同步地改变。
// 3. 通知 phaseManager，允许当前阶段做出反应并可能触发转换。
// 4. 事件被广播到外部侦听器并记录在游戏的历史记录中。
// eventType: 游戏事件的类型。
// payload: 事件的有效负载，必须是 Protobuf 消息。
func (em *eventManager) createAndProcess(eventType model.GameEventType, payload proto.Message) {
	anyPayload, err := anypb.New(payload)
	if err != nil {
		em.logger.Error("Failed to create anypb.Any for event payload", zap.Error(err))
		return
	}
	event := &model.GameEvent{
		Type:      eventType,
		Payload:   anyPayload,
		Timestamp: timestamppb.Now(),
	}

	// 步骤 1：同步应用状态更改。
	// 这对于确保游戏状态在任何其他逻辑运行之前保持一致至关重要。
	if err := eventhandler.ProcessEvent(em.engine.GameState, event); err != nil {
		em.logger.Error("Failed to apply event to game state", zap.Error(err), zap.String("type", event.Type.String()))
		// 即使处理程序失败，我们也会继续，以允许阶段和侦听器做出反应。
	}

	// 步骤 2：让当前阶段对事件做出反应。
	// 这现在由阶段管理器处理。
	em.engine.pm.handleEvent(event)

	// 步骤 3：将事件发布到外部侦听器并进行记录。
	// 这在状态更新后发生。
	select {
	case em.dispatchGameEvent <- event:
		// 同时在游戏状态中记录事件以供玩家查看
		em.engine.GameState.DayEvents = append(em.engine.GameState.DayEvents, event)
		em.engine.GameState.LoopEvents = append(em.engine.GameState.LoopEvents, event)
	default:
		em.logger.Warn("Game event channel full, dropping event", zap.String("eventType", event.Type.String()))
	}

	// TODO: 在状态完全更新后，在此处重新实现触发器逻辑。
	// em.engine.checkForTriggers(event)
}

// eventsChannel 返回游戏事件的出站通道。
// 返回值: 游戏事件的只读通道。
func (em *eventManager) eventsChannel() <-chan *model.GameEvent {
	return em.dispatchGameEvent
}

// close 关闭事件分发通道。
func (em *eventManager) close() {
	close(em.dispatchGameEvent)
}