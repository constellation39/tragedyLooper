package engine

import (
	model "tragedylooper/internal/game/proto/v1"

	"go.uber.org/zap"
)

// LocationGrid defines the 2x2 map layout.
var LocationGrid = map[model.LocationType]struct{ X, Y int }{
	model.LocationType_SHRINE:   {0, 0},
	model.LocationType_SCHOOL:   {1, 0},
	model.LocationType_HOSPITAL: {0, 1},
	model.LocationType_CITY:     {1, 1},
}

// resolveMovement processes all movement cards played in a turn.
func (ge *GameEngine) resolveMovement() {
	// Aggregate movement effects for each character
	charMovements := make(map[int32]struct{ H, V, D int })

	for _, card := range ge.GameState.PlayedCardsThisDay {
		targetCharID, isCharTarget := card.Target.(*model.Card_TargetCharacterId)
		if !isCharTarget {
			continue
		}

		movement, ok := charMovements[targetCharID.TargetCharacterId]
		if !ok {
			movement = struct{ H, V, D int }{0, 0, 0}
		}

		switch card.Config.Type {
		case model.CardType_MOVE_HORIZONTALLY:
			movement.H++
		case model.CardType_MOVE_VERTICALLY:
			movement.V++
		case model.CardType_MOVE_DIAGONALLY:
			movement.D++
		case model.CardType_FORBID_MOVEMENT:
			// If a forbid movement card is played, all other movement is cancelled.
			movement.H = -999 // Use a sentinel value to indicate cancellation
		}
		charMovements[targetCharID.TargetCharacterId] = movement
	}

	// Apply the calculated movements
	for charID, movement := range charMovements {
		if movement.H < 0 { // Movement was forbidden
			continue
		}

		char := ge.getCharacterByID(charID)
		if char == nil || !char.IsAlive {
			continue
		}

		finalH := movement.H
		finalV := movement.V

		// Diagonal movement counts as one horizontal and one vertical move.
		if movement.D > 0 {
			finalH += movement.D
			finalV += movement.D
		}

		// A combined H and V move becomes a diagonal move.
		if finalH > 0 && finalV > 0 {
			finalH--
			finalV--
			// Effectively, we are doing one diagonal move, and then any remaining H/V moves.
			ge.moveCharacter(char, 1, 1) // Diagonal
		}

		if finalH > 0 {
			ge.moveCharacter(char, finalH, 0) // Horizontal
		}
		if finalV > 0 {
			ge.moveCharacter(char, 0, finalV) // Vertical
		}
	}
}

func (ge *GameEngine) moveCharacter(char *model.Character, dx, dy int) {
	startPos, ok := LocationGrid[char.CurrentLocation]
	if !ok {
		ge.logger.Warn("character in unknown location", zap.String("char", char.Config.Name))
		return
	}

	// Calculate the new position, wrapping around the 2x2 grid.
	newX := (startPos.X + dx) % 2
	newY := (startPos.Y + dy) % 2

	var newLoc model.LocationType
	for loc, pos := range LocationGrid {
		if pos.X == newX && pos.Y == newY {
			newLoc = loc
			break
		}
	}

	if newLoc != model.LocationType_LOCATION_TYPE_UNSPECIFIED && newLoc != char.CurrentLocation {
		// Check for movement restrictions
		for _, rule := range char.Config.Rules {
			if smr, ok := rule.Effect.(*model.CharacterRule_SpecialMovementRule); ok {
				for _, restricted := range smr.SpecialMovementRule.RestrictedLocations {
					if restricted == newLoc {
						ge.logger.Info("character movement restricted", zap.String("char", char.Config.Name), zap.String("location", newLoc.String()))
						return // Movement forbidden
					}
				}
			}
		}

		char.CurrentLocation = newLoc
		ge.applyAndPublishEvent(model.GameEventType_CHARACTER_MOVED, &model.CharacterMovedEvent{
			CharacterId: char.Config.Id,
			NewLocation: newLoc,
		})
		ge.logger.Info("character moved", zap.String("char", char.Config.Name), zap.String("to", newLoc.String()))
	}
}

// resolveOtherCards handles non-movement cards.
func (ge *GameEngine) resolveOtherCards() {
	for _, card := range ge.GameState.PlayedCardsThisDay {
		switch card.Config.Type {
		case model.CardType_MOVE_HORIZONTALLY, model.CardType_MOVE_VERTICALLY, model.CardType_MOVE_DIAGONALLY, model.CardType_FORBID_MOVEMENT:
			continue // Already handled
		default:
			// TODO: Implement logic for other card types (Paranoia, Goodwill, etc.)
			ge.logger.Info("resolving other card", zap.String("card", card.Config.Name))
		}
	}
}
