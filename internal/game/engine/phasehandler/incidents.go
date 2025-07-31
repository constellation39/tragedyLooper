package phasehandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// IncidentsPhase 是检查和触发事件条件的阶段。
type IncidentsPhase struct{ basePhase }

// Type 返回阶段类型。
func (p *IncidentsPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_INCIDENTS }

// Enter 在阶段开始时调用。
func (p *IncidentsPhase) Enter(ge GameEngine) Phase {
	ge.TriggerIncidents()

	// 触发后，我们检查是否有任何待处理的选择。如果没有，我们可以继续。
	// 一个更健壮的系统可能会等待一个明确的信号，表明所有事件都已解决。
	return GetPhase(model.GamePhase_GAME_PHASE_DAY_END)
}

func init() {
	RegisterPhase(&IncidentsPhase{})
}

// HandleAction 处理事件阶段的行动，主要用于选择。
func (p *IncidentsPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if payload, ok := action.Payload.(*model.PlayerActionPayload_ChooseOption); ok {
		// 在这里，我们需要找到需要选择的原始效果
		// 并使用提供的选择重新应用它。这是一个复杂的问题。
		// 目前，我们将记录它并假设选择解决了某些问题。
		ge.Logger().Info("Player made a choice during IncidentsPhase", zap.Any("choice", payload))
		// 选择后，我们可能需要重新评估事件或其他条件。
		// 为简单起见，我们不在此处转换，假设游戏循环继续。
	}
	return nil
}
