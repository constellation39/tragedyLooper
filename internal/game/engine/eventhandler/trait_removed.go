package eventhandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

func init() {
	Register(model.GameEventType_TRAIT_REMOVED, &TraitRemovedHandler{})
}

// TraitRemovedHandler handles the TraitRemovedEvent.
type TraitRemovedHandler struct{}

// Handle removes a trait from a character.
func (h *TraitRemovedHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	e, ok := event.Payload.Payload.(*model.EventPayload_TraitRemoved)
	if !ok {
		return nil // Or handle error appropriately
	}

	state := ge.GetGameState()
	if char, ok := state.Characters[e.TraitRemoved.CharacterId]; ok {
		for i, t := range char.Traits {
			if t == e.TraitRemoved.Trait {
				char.Traits = append(char.Traits[:i], char.Traits[i+1:]...)
				return nil // Found and removed
			}
		}
	}
	return nil
}
