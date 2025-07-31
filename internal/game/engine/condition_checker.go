package engine

import (
	"fmt"
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

type conditionChecker struct {
	engine *GameEngine
}

func newConditionChecker(engine *GameEngine) *conditionChecker {
	return &conditionChecker{engine: engine}
}

// Check 根据当前游戏状态评估条件。
func (cc *conditionChecker) Check(gs *model.GameState, condition *model.Condition) (bool, error) {
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

func (cc *conditionChecker) checkCompoundCondition(gs *model.GameState, condition *model.CompoundCondition) (bool, error) {
	switch condition.Operator {
	case model.CompoundCondition_OPERATOR_AND:
		for _, sub := range condition.SubConditions {
			result, err := cc.Check(gs, sub)
			if err != nil || !result {
				return false, err
			}
		}
		return true, nil
	case model.CompoundCondition_OPERATOR_OR:
		for _, sub := range condition.SubConditions {
			result, err := cc.Check(gs, sub)
			if err == nil && result {
				return true, nil
			}
		}
		return false, nil
	case model.CompoundCondition_OPERATOR_NOT:
		if len(condition.SubConditions) != 1 {
			return false, fmt.Errorf("NOT operator requires exactly one sub-condition")
		}
		result, err := cc.Check(gs, condition.SubConditions[0])
		return !result, err
	default:
		return false, fmt.Errorf("unknown compound operator: %v", condition.Operator)
	}
}

func (cc *conditionChecker) getCharacter(gs *model.GameState, target *model.TargetSelector) (*model.Character, error) {
	// This is a simplified implementation. A full implementation would resolve the TargetSelector.
	// For now, we assume the target is always a specific character.
	char, ok := gs.Characters[target.CharacterId]
	if !ok {
		return nil, fmt.Errorf("character not found: %d", target.CharacterId)
	}
	return char, nil
}

func (cc *conditionChecker) checkStatCondition(gs *model.GameState, condition *model.StatCondition) (bool, error) {
	char, err := cc.getCharacter(gs, condition.Target)
	if err != nil {
		return false, fmt.Errorf("failed to get character for stat condition: %w", err)
	}

	var statValue int32
	switch condition.StatType {
	case model.StatCondition_STAT_TYPE_PARANOIA:
		statValue = char.Paranoia
	case model.StatCondition_STAT_TYPE_GOODWILL:
		statValue = char.Goodwill
	case model.StatCondition_STAT_TYPE_INTRIGUE:
		statValue = char.Intrigue
	default:
		return false, fmt.Errorf("unknown stat type: %v", condition.StatType)
	}

	switch condition.Comparator {
	case model.StatCondition_COMPARATOR_GREATER_THAN:
		return statValue > condition.Value, nil
	case model.StatCondition_COMPARATOR_LESS_THAN:
		return statValue < condition.Value, nil
	case model.StatCondition_COMPARATOR_EQUAL_TO:
		return statValue == condition.Value, nil
	case model.StatCondition_COMPARATOR_GREATER_THAN_OR_EQUAL:
		return statValue >= condition.Value, nil
	case model.StatCondition_COMPARATOR_LESS_THAN_OR_EQUAL:
		return statValue <= condition.Value, nil
	default:
		return false, fmt.Errorf("unknown comparator: %v", condition.Comparator)
	}
}

func (cc *conditionChecker) checkLocationCondition(gs *model.GameState, condition *model.LocationCondition) (bool, error) {
	char, err := cc.getCharacter(gs, condition.Target)
	if err != nil {
		return false, fmt.Errorf("failed to get character for location condition: %w", err)
	}

	isAtLocation := char.CurrentLocation == condition.Location
	return isAtLocation == condition.IsAtLocation, nil
}
