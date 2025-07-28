package engine

import (
	model "tragedylooper/internal/game/proto/v1"
)

func (ge *GameEngine) checkConditions(conditions []*model.Condition, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) bool {
	for _, condition := range conditions {
		if !ge.checkSingleCondition(condition, payload, choice) {
			return false
		}
	}
	return true
}

func (ge *GameEngine) checkSingleCondition(condition *model.Condition, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) bool {
	switch c := condition.ConditionType.(type) {
	case *model.Condition_StatCondition:
		return ge.checkStatCondition(c.StatCondition, payload, choice)
	case *model.Condition_LocationCondition:
		return ge.checkLocationCondition(c.LocationCondition, payload, choice)
		// Add other condition checks here
	}
	return false
}

func (ge *GameEngine) checkStatCondition(sc *model.StatCondition, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) bool {
	targetIDs, err := ge.resolveSelectorToCharacters(ge.GameState, sc.Target, payload, choice)
	if err != nil {
		// Log the error, maybe?
		return false
	}

	// For stat conditions, we usually expect a single target.
	// If the selector resolves to multiple, the condition is false unless all of them meet it.
	if len(targetIDs) == 0 {
		return false // Or true, depending on desired logic for empty sets
	}

	for _, charID := range targetIDs {
		char, ok := ge.GameState.Characters[charID]
		if !ok {
			continue // Or return false
		}

		var statValue int32
		switch sc.StatType {
		case model.StatCondition_PARANOIA:
			statValue = char.Paranoia
		case model.StatCondition_GOODWILL:
			statValue = char.Goodwill
		case model.StatCondition_INTRIGUE:
			statValue = char.Intrigue
		default:
			return false // Unknown stat type
		}

		conditionMet := false
		switch sc.Comparator {
		case model.StatCondition_GREATER_THAN:
			conditionMet = statValue > sc.Value
		case model.StatCondition_LESS_THAN:
			conditionMet = statValue < sc.Value
		case model.StatCondition_EQUAL_TO:
			conditionMet = statValue == sc.Value
		case model.StatCondition_GREATER_THAN_OR_EQUAL:
			conditionMet = statValue >= sc.Value
		case model.StatCondition_LESS_THAN_OR_EQUAL:
			conditionMet = statValue <= sc.Value
		}

		// If any character does not meet the condition, the overall condition is false.
		if !conditionMet {
			return false
		}
	}

	// If we get here, all targeted characters met the condition.
	return true
}

func (ge *GameEngine) checkLocationCondition(lc *model.LocationCondition, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) bool {
	targetIDs, err := ge.resolveSelectorToCharacters(ge.GameState, lc.Target, payload, choice)
	if err != nil {
		return false
	}

	if len(targetIDs) == 0 {
		return false
	}

	for _, charID := range targetIDs {
		char, ok := ge.GameState.Characters[charID]
		if !ok {
			continue
		}

		atLocation := char.CurrentLocation == lc.Location
		if lc.IsAtLocation && !atLocation {
			return false
		}
		if !lc.IsAtLocation && atLocation { // For checking if NOT at a location
			return false
		}

		if lc.IsAlone || lc.NotAlone {
			numOthersAtLocation := 0
			for otherID, otherChar := range ge.GameState.Characters {
				if otherID != charID && otherChar.CurrentLocation == lc.Location {
					numOthersAtLocation++
				}
			}
			if lc.IsAlone && numOthersAtLocation > 0 {
				return false
			}
			if lc.NotAlone && numOthersAtLocation == 0 {
				return false
			}
		}
	}

	return true
}

func (ge *GameEngine) checkGameEndConditions() (bool, model.PlayerRole) {
	// This logic needs to be updated based on the script's end conditions.
	// The following is placeholder logic.

	// Example: Check if max loops are reached.
	// if ge.GameState.CurrentLoop > ge.GameState.Script.LoopCount {
	// 	return true, model.PlayerRole_MASTERMIND // Or based on who has more points, etc.
	// }

	return false, model.PlayerRole_PLAYER_ROLE_UNSPECIFIED
}
