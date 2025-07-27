package engine

import (
	"fmt"
	model "tragedylooper/internal/game/proto/v1"
)

func (ge *GameEngine) checkConditions(conditions []*model.Condition) bool {
	for _, condition := range conditions {
		switch c := condition.ConditionType.(type) {
		case *model.Condition_StatCondition:
			sc := c.StatCondition
			char, ok := ge.GameState.Characters[sc.CharacterId]
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
				if !(statValue > sc.Value) {
					return false
				}
			case model.Operator_OPERATOR_LESS_THAN:
				if !(statValue < sc.Value) {
					return false
				}
			case model.Operator_OPERATOR_EQUAL_TO:
				if !(statValue == sc.Value) {
					return false
				}
			case model.Operator_OPERATOR_GREATER_THAN_OR_EQUAL_TO:
				if !(statValue >= sc.Value) {
					return false
				}
			case model.Operator_OPERATOR_LESS_THAN_OR_EQUAL_TO:
				if !(statValue <= sc.Value) {
					return false
				}
			}

		case *model.Condition_LocationCondition:
			lc := c.LocationCondition
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
		}
	}
	return true
}

func (ge *GameEngine) checkGameEndConditions() (bool, model.PlayerRole) {
	// Check for protagonist win conditions
	for _, wc := range ge.GameState.Script.WinConditions {
		switch wc.Type {
		case model.GameEndConditionType_ALL_TRAGEDIES_PREVENTED:
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
		switch lc.Type {
		case model.GameEndConditionType_A_TRAGEDY_OCCURS:
			// This is checked within the incident phase, so we just need to see if a tragedy has occurred.
			for _, occurred := range ge.GameState.TragedyOccurred {
				if occurred {
					return true, model.PlayerRole_PLAYER_ROLE_MASTERMIND
				}
			}
		}
	}

	return false, model.PlayerRole_PLAYER_ROLE_UNSPECIFIED
}
