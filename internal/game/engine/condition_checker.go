package engine

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

// CheckCondition evaluates a condition against the current game state.
func (ge *GameEngine) CheckCondition(condition *model.Condition) (bool, error) {
	if condition == nil {
		return true, nil // A nil condition is considered true
	}

	switch c := condition.ConditionType.(type) {
	case *model.Condition_StatCondition:
		return ge.checkStatCondition(c.StatCondition)
	case *model.Condition_LocationCondition:
		return ge.checkLocationCondition(c.LocationCondition)
	// Add other condition checks here as they are implemented
	case *model.Condition_CompoundCondition:
		return ge.checkCompoundCondition(c.CompoundCondition)
	default:
		return false, fmt.Errorf("unhandled condition type: %T", c)
	}
}

func (ge *GameEngine) checkCompoundCondition(condition *model.CompoundCondition) (bool, error) {
	switch condition.Operator {
	case model.CompoundCondition_AND:
		for _, sub := range condition.SubConditions {
			result, err := ge.CheckCondition(sub)
			if err != nil || !result {
				return false, err
			}
		}
		return true, nil
	case model.CompoundCondition_OR:
		for _, sub := range condition.SubConditions {
			result, err := ge.CheckCondition(sub)
			if err == nil && result {
				return true, nil
			}
		}
		return false, nil
	case model.CompoundCondition_NOT:
		if len(condition.SubConditions) != 1 {
			return false, fmt.Errorf("NOT operator requires exactly one sub-condition")
		}
		result, err := ge.CheckCondition(condition.SubConditions[0])
		return !result, err
	default:
		return false, fmt.Errorf("unknown compound operator: %v", condition.Operator)
	}
}

func (ge *GameEngine) checkStatCondition(condition *model.StatCondition) (bool, error) {
	// This is a simplified implementation. A full implementation would need to resolve the TargetSelector.
	// For now, we'll assume the target is always a specific character.
	char, ok := ge.GameState.Characters[condition.Target.CharacterId]
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

func (ge *GameEngine) checkLocationCondition(condition *model.LocationCondition) (bool, error) {
	// Simplified implementation
	char, ok := ge.GameState.Characters[condition.Target.CharacterId]
	if !ok {
		return false, fmt.Errorf("character not found in location condition: %d", condition.Target.CharacterId)
	}

	isAtLocation := char.CurrentLocation == condition.Location
	return isAtLocation == condition.IsAtLocation, nil
}
