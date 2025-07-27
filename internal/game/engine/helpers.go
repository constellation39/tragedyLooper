package engine

import (
	model "tragedylooper/internal/game/proto/v1"
)

func (ge *GameEngine) checkConditions(conditions []*model.Condition) bool {
	for _, condition := range conditions {
		if !ge.checkSingleCondition(condition) {
			return false
		}
	}
	return true
}

func (ge *GameEngine) checkSingleCondition(condition *model.Condition) bool {
	switch c := condition.ConditionType.(type) {
	case *model.Condition_StatCondition:
		return ge.checkStatCondition(c.StatCondition)
	case *model.Condition_LocationCondition:
		return ge.checkLocationCondition(c.LocationCondition)
	}
	return false
}

func (ge *GameEngine) checkStatCondition(sc *model.StatCondition) bool {
	char, ok := ge.GameState.Characters[sc.CharacterId]
	if !ok {
		return false
	}

	var statValue int32
	switch sc.StatType {
	case model.StatCondition_PARANOIA_STAT:
		statValue = char.Paranoia
	case model.StatCondition_GOODWILL_STAT:
		statValue = char.Goodwill
	case model.StatCondition_INTRIGUE_STAT:
		statValue = char.Intrigue
	}

	switch sc.Comparator {
	case model.StatCondition_GREATER_THAN:
		return statValue > sc.Value
	case model.StatCondition_LESS_THAN:
		return statValue < sc.Value
	case model.StatCondition_EQUAL_TO:
		return statValue == sc.Value
	case model.StatCondition_GREATER_THAN_OR_EQUAL:
		return statValue >= sc.Value
	case model.StatCondition_LESS_THAN_OR_EQUAL:
		return statValue <= sc.Value
	}
	return false
}

func (ge *GameEngine) checkLocationCondition(lc *model.LocationCondition) bool {
	char, ok := ge.GameState.Characters[lc.CharacterId]
	if !ok {
		return false
	}

	if char.CurrentLocation != lc.Location {
		return false
	}

	if lc.IsAlone {
		for _, otherChar := range ge.GameState.Characters {
			if otherChar.Id != char.Id && otherChar.CurrentLocation == char.CurrentLocation {
				return false
			}
		}
	}
	return true
}

func (ge *GameEngine) checkGameEndConditions() (bool, model.PlayerRole) {
	// Check for protagonist win conditions
	for _, wc := range ge.GameState.Script.WinConditions {
		if wc.Type == model.GameEndCondition_ALL_TRAGEDIES_PREVENTED {
			allPrevented := true
			for _, prevented := range ge.GameState.PreventedTragedies {
				if !prevented {
					allPrevented = false
					break
				}
			}
			if allPrevented {
				return true, model.PlayerRole_PROTAGONIST
			}
		}
	}

	// Check for mastermind win conditions
	for _, lc := range ge.GameState.Script.LoseConditions {
		if lc.Type == model.GameEndCondition_SPECIFIC_TRAGEDY_TRIGGERED {
			// This is checked within the incident phase, so we just need to see if a tragedy has occurred.
			for _, active := range ge.GameState.ActiveTragedies {
				if !active {
					return true, model.PlayerRole_MASTERMIND
				}
			}
		}
	}

	return false, model.PlayerRole_PLAYER_ROLE_UNSPECIFIED
}
