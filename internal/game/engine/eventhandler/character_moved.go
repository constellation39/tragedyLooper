package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

func init() {
	Register(model.GameEventType_CHARACTER_MOVED, &CharacterMovedHandler{})
}

// CharacterMovedHandler handles the CharacterMovedEvent.
type CharacterMovedHandler struct{}

// Handle updates the character's location in the game state.
func (h *CharacterMovedHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	e, ok := event.Payload.Payload.(*model.EventPayload_CharacterMoved)
	if !ok {
		return nil // Or handle error appropriately
	}

	state := ge.GetGameState()
	if char, ok := state.Characters[e.CharacterMoved.CharacterId]; ok {
		char.CurrentLocation = e.CharacterMoved.NewLocation
		ge.Logger().Info("character moved", zap.String("char", char.Config.Name), zap.String("to", e.CharacterMoved.NewLocation.String()))
	}
	return nil
}
