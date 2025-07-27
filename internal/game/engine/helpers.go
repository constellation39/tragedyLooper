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
	charID, ok := ge.characterNameToID[sc.CharacterId]
	if !ok {
		return false
	}
	char, ok := ge.GameState.Characters[charID]
	if !ok {
		return false
	}

	var statValue int32
	switch sc.Stat {
	case model.Stat_STAT_PARANOIA:
		statValue = char.Paranoia
	case model.Stat_STAT_GOODWILL:
		statValue = char.Goodwill
	case model.Stat_STAT_INTRIGUE:
		statValue = char.Intrigue
	}

	switch sc.Operator {
	case model.Operator_OPERATOR_GREATER_THAN:
		return statValue > sc.Value
	case model.Operator_OPERATOR_LESS_THAN:
		return statValue < sc.Value
	case model.Operator_OPERATOR_EQUAL_TO:
		return statValue == sc.Value
	case model.Operator_OPERATOR_GREATER_THAN_OR_EQUAL_TO:
		return statValue >= sc.Value
	case model.Operator_OPERATOR_LESS_THAN_OR_EQUAL_TO:
		return statValue <= sc.Value
	}
	return false
}

func (ge *GameEngine) checkLocationCondition(lc *model.LocationCondition) bool {
	charID, ok := ge.characterNameToID[lc.CharacterId]
	if !ok {
		return false
	}
	char, ok := ge.GameState.Characters[charID]
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
		if wc.Type == model.GameEndConditionType_ALL_TRAGEDIES_PREVENTED {
			allPrevented := true
			for _, prevented := range ge.GameState.PreventedTragedies {
				if !prevented {
					allPrevented = false
					break
				}
			}
			if allPrevented {
				return true, model.PlayerRole_PLAYER_ROLE_PROTAGONIST
			}
		}
	}

	// Check for mastermind win conditions
	for _, lc := range ge.GameState.Script.LoseConditions {
		if lc.Type == model.GameEndConditionType_A_TRAGEDY_OCCURS {
			// This is checked within the incident phase, so we just need to see if a tragedy has occurred.
			for _, active := range ge.GameState.ActiveTragedies {
				if !active {
					return true, model.PlayerRole_PLAYER_ROLE_MASTERMIND
				}
			}
		}
	}

	return false, model.PlayerRole_PLAYER_ROLE_UNSPECIFIED
}
