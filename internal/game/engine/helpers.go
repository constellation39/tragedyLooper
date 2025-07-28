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

	if len(targetIDs) == 0 {
		return false // Or true, depending on desired logic for empty sets
	}

	for _, charID := range targetIDs {
		char, ok := ge.GameState.Characters[charID]
		if !ok {
			continue // Or return false
		}

		statValue, ok := getStatValue(char, sc.StatType)
		if !ok {
			return false // Unknown stat type
		}

		if !compareStat(statValue, sc.Value, sc.Comparator) {
			return false
		}
	}

	return true
}

func getStatValue(char *model.Character, statType model.StatCondition_StatType) (int32, bool) {
	switch statType {
	case model.StatCondition_PARANOIA:
		return char.Paranoia, true
	case model.StatCondition_GOODWILL:
		return char.Goodwill, true
	case model.StatCondition_INTRIGUE:
		return char.Intrigue, true
	default:
		return 0, false
	}
}

func compareStat(statValue, conditionValue int32, comparator model.StatCondition_Comparator) bool {
	switch comparator {
	case model.StatCondition_GREATER_THAN:
		return statValue > conditionValue
	case model.StatCondition_LESS_THAN:
		return statValue < conditionValue
	case model.StatCondition_EQUAL_TO:
		return statValue == conditionValue
	case model.StatCondition_GREATER_THAN_OR_EQUAL:
		return statValue >= conditionValue
	case model.StatCondition_LESS_THAN_OR_EQUAL:
		return statValue <= conditionValue
	}
	return false
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
		if lc.IsAtLocation != atLocation {
			return false
		}

		if !ge.checkCharacterIsolation(lc, char) {
			return false
		}
	}

	return true
}

func (ge *GameEngine) checkCharacterIsolation(lc *model.LocationCondition, char *model.Character) bool {
	if !lc.IsAlone && !lc.NotAlone {
		return true // No isolation condition to check
	}

	numOthersAtLocation := 0
	for otherID, otherChar := range ge.GameState.Characters {
		if otherID != char.Config.Id && otherChar.CurrentLocation == lc.Location {
			numOthersAtLocation++
		}
	}

	if lc.IsAlone && numOthersAtLocation > 0 {
		return false
	}

	if lc.NotAlone && numOthersAtLocation == 0 {
		return false
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
