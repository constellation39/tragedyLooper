package engine

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

type ConditionChecker struct {
	engine *GameEngine
}

func NewConditionChecker(engine *GameEngine) *ConditionChecker {
	return &ConditionChecker{engine: engine}
}

// Check 根据当前游戏状态评估条件。
func (cc *ConditionChecker) Check(gs *model.GameState, condition *model.Condition) (bool, error) {
	if condition == nil {
		return true, nil // nil 条件被视为 true
	}

	switch c := condition.ConditionType.(type) {
	case *model.Condition_StatCondition:
		return cc.checkStatCondition(gs, c.StatCondition)
	case *model.Condition_LocationCondition:
		return cc.checkLocationCondition(gs, c.LocationCondition)
	// 在此处添加其他条件检查
	case *model.Condition_CompoundCondition:
		return cc.checkCompoundCondition(gs, c.CompoundCondition)
	default:
		return false, fmt.Errorf("unhandled condition type: %T", c)
	}
}

func (cc *ConditionChecker) checkCompoundCondition(gs *model.GameState, condition *model.CompoundCondition) (bool, error) {
	switch condition.Operator {
	case model.CompoundCondition_AND:
		for _, sub := range condition.SubConditions {
			result, err := cc.Check(gs, sub)
			if err != nil || !result {
				return false, err
			}
		}
		return true, nil
	case model.CompoundCondition_OR:
		for _, sub := range condition.SubConditions {
			result, err := cc.Check(gs, sub)
			if err == nil && result {
				return true, nil
			}
		}
		return false, nil
	case model.CompoundCondition_NOT:
		if len(condition.SubConditions) != 1 {
			return false, fmt.Errorf("NOT operator requires exactly one sub-condition")
		}
		result, err := cc.Check(gs, condition.SubConditions[0])
		return !result, err
	default:
		return false, fmt.Errorf("unknown compound operator: %v", condition.Operator)
	}
}

func (cc *ConditionChecker) checkStatCondition(gs *model.GameState, condition *model.StatCondition) (bool, error) {
	// 这是一个简化的实现。完整的实现需要解析 TargetSelector。
	// 目前，我们假设目标始终是特定角色。
	char, ok := gs.Characters[condition.Target.CharacterId]
	if !ok {
		return false, fmt.Errorf("character not found in stat condition: %d", condition.Target.CharacterId)
	}

	var statValue int32
	switch condition.StatType {
	case model.StatCondition_PARANOIA:
		statValue = char.Paranoia
	case model.StatCondition_GOODWILL:
		statValue = char.Goodwill
	case model.StatCondition_INTRIGUE:
		statValue = char.Intrigue
	default:
		return false, fmt.Errorf("unknown stat type: %v", condition.StatType)
	}

	switch condition.Comparator {
	case model.StatCondition_GREATER_THAN:
		return statValue > condition.Value, nil
	case model.StatCondition_LESS_THAN:
		return statValue < condition.Value, nil
	case model.StatCondition_EQUAL_TO:
		return statValue == condition.Value, nil
	case model.StatCondition_GREATER_THAN_OR_EQUAL:
		return statValue >= condition.Value, nil
	case model.StatCondition_LESS_THAN_OR_EQUAL:
		return statValue <= condition.Value, nil
	default:
		return false, fmt.Errorf("unknown comparator: %v", condition.Comparator)
	}
}

func (cc *ConditionChecker) checkLocationCondition(gs *model.GameState, condition *model.LocationCondition) (bool, error) {
	// 简化实现
	char, ok := gs.Characters[condition.Target.CharacterId]
	if !ok {
		return false, fmt.Errorf("character not found in location condition: %d", condition.Target.CharacterId)
	}

	isAtLocation := char.CurrentLocation == condition.Location
	return isAtLocation == condition.IsAtLocation, nil
}