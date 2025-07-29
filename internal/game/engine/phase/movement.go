package phase

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- MovementPhase ---
type MovementPhase struct{ basePhase }

func (p *MovementPhase) Type() model.GamePhase { return model.GamePhase_MOVEMENT }

func (p *MovementPhase) Enter(ge GameEngine) Phase {
	// Here you would implement the logic to determine which characters move
	// and where they move to, based on the game state and played cards.
	//
	// For now, this is a placeholder.
	//
	// Example:
	// state := ge.GetGameState()
	// for _, char := range state.Characters {
	//     if shouldMove(char) {
	//         dx, dy := calculateMovement(char)
	//         ge.MoveCharacter(char, dx, dy) // Assuming MoveCharacter is still on GameEngine
	//     }
	// }

	// After movement is resolved, we might move to the card effects phase.
	return &CardEffectsPhase{}
}
