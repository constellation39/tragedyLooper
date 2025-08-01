package phasehandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// DayEndPhase 是执行天末检查的阶段。
type DayEndPhase struct {
	BasePhase
}

// Type 返回阶段类型。
func (p *DayEndPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_DAY_END }

// Enter 在阶段开始时调用。
func (p *DayEndPhase) Enter(ge GameEngine) {
	logger := ge.Logger().Named("DayEndPhase")
	script := ge.GetGameRepo().GetScript()

	// 1. 检查循环失败条件
	for _, endCond := range script.LoseConditions {
		if endCond.Type == model.EndConditionType_END_CONDITION_TYPE_PROTAGONIST_GUESS_FAIL {
			for _, req := range endCond.Requirements {
				met, err := ge.CheckCondition(req)
				if err != nil {
					logger.Error("Error checking loop loss condition", zap.Error(err))
					continue
				}
				if met {
					logger.Info("Loop loss condition met", zap.String("description", endCond.Description))
					ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_LOOP_LOSS, &model.EventPayload{})
					return
				}
			}
		}
	}

	// 2. 检查主角胜利条件（例如，所有失败条件都已阻止）
	// 这个逻辑可能很复杂。一个简单的版本是检查作为失败条件一部分的所有事件是否都已阻止。
	// 目前，我们假设一个简单的检查。

	// 3. 如果没有胜利/失败，检查是否是循环的最后一天
	if ge.GetGameState().CurrentDay >= ge.GetGameState().DaysPerLoop {
		logger.Info("End of loop reached by day count")
		return
	}

	// 4. 否则，进入下一天
	logger.Info("Proceeding to next day")
}

func init() {
	RegisterPhase(&DayEndPhase{})
}
