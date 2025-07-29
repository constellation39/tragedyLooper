package phase

import (
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// --- CardEffectsPhase ---
type CardEffectsPhase struct{ basePhase }

func (p *CardEffectsPhase) Type() model.GamePhase { return model.GamePhase_CARD_RESOLVE }

func (p *CardEffectsPhase) Enter(ge GameEngine) Phase {
	resolver := NewMovementResolver()
	charMovements := resolver.CalculateMovements(ge.GetGameState().PlayedCardsThisDay)

	// Apply the calculated movements
	for charID, movement := range charMovements {
		if movement.Forbidden {
			continue
		}

		char := ge.GetCharacterByID(charID)
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
			ge.MoveCharacter(char, 1, 1) // Diagonal
		}

		if finalH > 0 {
			ge.MoveCharacter(char, finalH, 0) // Horizontal
		}
		if finalV > 0 {
			ge.MoveCharacter(char, 0, finalV) // Vertical
		}
	}

	resolveOtherCards(ge)

	// After card effects are resolved, we might move to the abilities phase.
	return &AbilitiesPhase{}
}

// CharacterMovement holds the calculated movement vectors for a character.
type CharacterMovement struct {
	H         int
	V         int
	D         int
	Forbidden bool
}

// MovementResolver calculates character movements based on played cards.
type MovementResolver struct{}

// NewMovementResolver creates a new MovementResolver.
func NewMovementResolver() *MovementResolver {
	return &MovementResolver{}
}

// CalculateMovements aggregates movement effects for each character from the played cards.
func (mr *MovementResolver) CalculateMovements(playedCards map[int32]*model.Card) map[int32]CharacterMovement {
	charMovements := make(map[int32]CharacterMovement)

	for _, card := range playedCards {
		targetCharID, isCharTarget := card.Target.(*model.Card_TargetCharacterId)
		if !isCharTarget {
			continue
		}

		movement := charMovements[targetCharID.TargetCharacterId]
		if movement.Forbidden {
			continue // Movement is already forbidden, no further calculations needed.
		}

		switch card.Config.Type {
		case model.CardType_MOVE_HORIZONTALLY:
			movement.H++
		case model.CardType_MOVE_VERTICALLY:
			movement.V++
		case model.CardType_MOVE_DIAGONALLY:
			movement.D++
		case model.CardType_FORBID_MOVEMENT:
			movement = CharacterMovement{Forbidden: true} // Cancel all movement.
		}
		charMovements[targetCharID.TargetCharacterId] = movement
	}
	return charMovements
}

// resolveMovement processes all movement cards played in a turn.
func resolveMovement(ge GameEngine) {
	resolver := NewMovementResolver()
	charMovements := resolver.CalculateMovements(ge.GetGameState().PlayedCardsThisDay)

	// Apply the calculated movements
	for charID, movement := range charMovements {
		if movement.Forbidden {
			continue
		}

		char := ge.GetCharacterByID(charID)
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
			ge.MoveCharacter(char, 1, 1) // Diagonal
			}

			if finalH > 0 {
				ge.MoveCharacter(char, finalH, 0) // Horizontal
			}
			if finalV > 0 {
				ge.MoveCharacter(char, 0, finalV) // Vertical
		}
	}
}


// resolveOtherCards handles non-movement cards.
func resolveOtherCards(ge GameEngine) {
	for _, card := range ge.GetGameState().PlayedCardsThisDay {
		switch card.Config.Type {
		case model.CardType_MOVE_HORIZONTALLY, model.CardType_MOVE_VERTICALLY, model.CardType_MOVE_DIAGONALLY, model.CardType_FORBID_MOVEMENT:
			continue // Already handled
		default:
			// TODO: Implement logic for other card types (Paranoia, Goodwill, etc.)
			ge.Logger().Info("resolving other card", zap.String("card", card.Config.Name))
		}
	}
}
