package condition

import (
	"fmt"

	"github.com/constellation39/tragedyLooper/internal/game/engine/target"
	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// Checker is responsible for evaluating conditions against the game state.
// It depends on a TargetResolver to correctly identify characters and locations from selectors.
type Checker struct {
	Resolver *target.Resolver
}

// NewChecker creates a new condition checker.
func NewChecker(resolver *target.Resolver) *Checker {
	return &Checker{Resolver: resolver}
}

// Check evaluates a condition against the current game state.
func (c *Checker) Check(gs *v1.GameState, condition *v1.Condition) (bool, error) {
	if condition == nil {
		return true, nil // A nil condition is always considered true.
	}

	switch cond := condition.ConditionType.(type) {
	case *v1.Condition_StatCondition:
		return c.checkStatCondition(gs, cond.StatCondition)
	case *v1.Condition_LocationCondition:
		return c.checkLocationCondition(gs, cond.LocationCondition)
	case *v1.Condition_RoleCondition:
		return c.checkRoleCondition(gs, cond.RoleCondition)
	case *v1.Condition_TraitCondition:
		return c.checkTraitCondition(gs, cond.TraitCondition)
	case *v1.Condition_DayCondition:
		return c.checkDayCondition(gs, cond.DayCondition)
	case *v1.Condition_PhaseCondition:
		return c.checkPhaseCondition(gs, cond.PhaseCondition)
	case *v1.Condition_CompoundCondition:
		return c.checkCompoundCondition(gs, cond.CompoundCondition)
	case *v1.Condition_EventHistoryCondition:
		return c.checkEventHistoryCondition(gs, cond.EventHistoryCondition)
	case *v1.Condition_LocationCharacterCountCondition:
		return c.checkLocationCharacterCountCondition(gs, cond.LocationCharacterCountCondition)
	default:
		return false, fmt.Errorf("unhandled condition type: %T", cond)
	}
}

func (c *Checker) checkCompoundCondition(gs *v1.GameState, condition *v1.CompoundCondition) (bool, error) {
	switch condition.Operator {
	case v1.CompoundCondition_OPERATOR_AND:
		if len(condition.SubConditions) == 0 {
			return true, nil // Vacuously true
		}
		for _, sub := range condition.SubConditions {
			result, err := c.Check(gs, sub)
			if err != nil || !result {
				return false, err
			}
		}
		return true, nil
	case v1.CompoundCondition_OPERATOR_OR:
		if len(condition.SubConditions) == 0 {
			return false, nil // Vacuously false
		}
		for _, sub := range condition.SubConditions {
			result, err := c.Check(gs, sub)
			if err == nil && result {
				return true, nil
			}
		}
		return false, nil
	case v1.CompoundCondition_OPERATOR_NOT:
		if len(condition.SubConditions) != 1 {
			return false, fmt.Errorf("NOT operator requires exactly one sub-condition, got %d", len(condition.SubConditions))
		}
		result, err := c.Check(gs, condition.SubConditions[0])
		return !result, err
	default:
		return false, fmt.Errorf("unknown compound operator: %v", condition.Operator)
	}
}

func (c *Checker) checkStatCondition(gs *v1.GameState, condition *v1.StatCondition) (bool, error) {
	// Per documentation, if a selector matches multiple characters, the condition is true if *any* of them satisfy it.
	chars, err := c.Resolver.ResolveCharacters(gs, condition.Target)
	if err != nil {
		return false, fmt.Errorf("failed to resolve target for stat condition: %w", err)
	}

	valueToCompare, err := c.resolveStatValue(gs, condition)
	if err != nil {
		return false, err
	}

	for _, char := range chars {
		statValue := getStat(char, condition.StatType)
		if compare(statValue, valueToCompare, condition.Comparator) {
			return true, nil
		}
	}
	return false, nil
}

func (c *Checker) resolveStatValue(gs *v1.GameState, condition *v1.StatCondition) (int32, error) {
	if condition.TargetToCompare != nil {
		// We are comparing against another character's stat.
		otherChars, err := c.Resolver.ResolveCharacters(gs, condition.TargetToCompare)
		if err != nil {
			return 0, fmt.Errorf("failed to resolve target_to_compare: %w", err)
		}
		if len(otherChars) != 1 {
			return 0, fmt.Errorf("target_to_compare must resolve to exactly one character, got %d", len(otherChars))
		}
		return getStat(otherChars[0], condition.StatType), nil
	}
	// We are comparing against a fixed value.
	return condition.Value, nil
}

func (c *Checker) checkLocationCondition(gs *v1.GameState, condition *v1.LocationCondition) (bool, error) {
	chars, err := c.Resolver.ResolveCharacters(gs, condition.Target)
	if err != nil {
		return false, fmt.Errorf("failed to resolve target for location condition: %w", err)
	}

	for _, char := range chars {
		if char.Location == condition.Location {
			return true, nil
		}
	}
	return false, nil
}

func (c *Checker) checkLocationCharacterCountCondition(gs *v1.GameState, condition *v1.LocationCharacterCountCondition) (bool, error) {
	count := 0
	for _, char := range gs.Characters {
		if char.Location == condition.Location {
			count++
		}
	}
	return compare(int32(count), condition.Count, condition.Comparator), nil
}

func (c *Checker) checkRoleCondition(gs *v1.GameState, condition *v1.RoleCondition) (bool, error) {
	chars, err := c.Resolver.ResolveCharacters(gs, condition.Target)
	if err != nil {
		return false, fmt.Errorf("failed to resolve target for role condition: %w", err)
	}

	for _, char := range chars {
		if char.Role.Id == condition.RoleId {
			return true, nil
		}
	}
	return false, nil
}

func (c *Checker) checkTraitCondition(gs *v1.GameState, condition *v1.TraitCondition) (bool, error) {
	chars, err := c.Resolver.ResolveCharacters(gs, condition.Target)
	if err != nil {
		return false, fmt.Errorf("failed to resolve target for trait condition: %w", err)
	}

	for _, char := range chars {
		for _, trait := range char.Traits {
			if trait == condition.Trait {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Checker) checkDayCondition(gs *v1.GameState, condition *v1.DayCondition) (bool, error) {
	return compare(gs.Day, condition.Day, condition.Comparator), nil
}

func (c *Checker) checkPhaseCondition(gs *v1.GameState, condition *v1.PhaseCondition) (bool, error) {
	// This requires a defined order for phases.
	// Assuming the enum values represent the order.
	return compare(int32(gs.Phase), int32(condition.Phase), condition.Comparator), nil
}

func (c *Checker) checkEventHistoryCondition(gs *v1.GameState, condition *v1.EventHistoryCondition) (bool, error) {
	// This is a complex condition that requires iterating through the event log.
	// The implementation will depend on how the event log is structured.
	// Placeholder implementation:
	return false, fmt.Errorf("event history condition not yet implemented")
}

// getStat is a helper to retrieve a stat value from a character.
func getStat(char *v1.Character, statType v1.StatType) int32 {
	switch statType {
	case v1.StatType_STAT_TYPE_PARANOIA:
		return char.Paranoia
	case v1.StatType_STAT_TYPE_INTRIGUE:
		return char.Intrigue
	// Note: Goodwill is on the Role, not the Character sheet itself.
	case v1.StatType_STAT_TYPE_GOODWILL:
		return char.Role.Goodwill
	default:
		return 0 // Or handle as an error
	}
}

// compare is a generic comparison helper for different numeric types.
func compare[T int32 | v1.GamePhase](a, b T, comparator v1.Comparator) bool {
	switch comparator {
	case v1.Comparator_EQUAL_TO:
		return a == b
	case v1.Comparator_NOT_EQUAL_TO:
		return a != b
	case v1.Comparator_GREATER_THAN:
		return a > b
	case v1.Comparator_LESS_THAN:
		return a < b
	case v1.Comparator_GREATER_THAN_OR_EQUAL_TO:
		return a >= b
	case v1.Comparator_LESS_THAN_OR_EQUAL_TO:
		return a <= b
	default:
		return false // Or handle as an error
	}
}
