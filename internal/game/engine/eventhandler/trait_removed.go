package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register(model.GameEventType_TRAIT_REMOVED, &TraitRemovedHandler{})
}

// TraitRemovedHandler handles the TraitRemovedEvent.
type TraitRemovedHandler struct{}

// Handle removes a trait from a character.
func (h *TraitRemovedHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	var e model.TraitRemovedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		return err
	}

	if char, ok := state.Characters[e.CharacterId]; ok {
		for i, t := range char.Traits {
			if t == e.Trait {
				char.Traits = append(char.Traits[:i], char.Traits[i+1:]...)
				return nil // Found and removed
			}
		}
	}
	return nil
}
