package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register(model.GameEventType_TRAIT_ADDED, &TraitAddedHandler{})
}

// TraitAddedHandler handles the TraitAddedEvent.
type TraitAddedHandler struct{}

// Handle adds a trait to a character if it doesn't exist yet.
func (h *TraitAddedHandler) Handle(state *model.GameState, event *model.EventPayload) error {
	e, ok := event.Payload.(*model.EventPayload_TraitAdded)
	if !ok {
		return nil // Or handle error appropriately
	}

	if char, ok := state.Characters[e.TraitAdded.CharacterId]; ok {
		// Avoid duplicates
		for _, t := range char.Traits {
			if t == e.TraitAdded.Trait {
				return nil // Already exists, not an error
			}
		}
		char.Traits = append(char.Traits, e.TraitAdded.Trait)
	}
	return nil
}
