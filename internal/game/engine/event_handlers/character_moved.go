package handlers

import (
	model "tragedylooper/internal/game/proto/v1"
)

// CharacterMovedHandler handles the CharacterMovedEvent.
type CharacterMovedHandler struct{}

// Handle updates the character's location in the game state.
func (h *CharacterMovedHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	var e model.CharacterMovedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		return err
	}

	if char, ok := state.Characters[e.CharacterId]; ok {
		char.CurrentLocation = e.NewLocation
	}
	return nil
}
