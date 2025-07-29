package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register(model.GameEventType_CHARACTER_MOVED, &CharacterMovedHandler{})
}

// CharacterMovedHandler handles the CharacterMovedEvent.
type CharacterMovedHandler struct{}

// Handle updates the character's location in the game state.
func (h *CharacterMovedHandler) Handle(state *model.GameState, event *model.EventPayload) error {
	e := event.Payload.GetCharacterMoved()
	if e == nil {
		return nil // Or handle error appropriately
	}

	if char, ok := state.Characters[e.CharacterId]; ok {
		char.CurrentLocation = e.NewLocation
	}
	return nil
}
