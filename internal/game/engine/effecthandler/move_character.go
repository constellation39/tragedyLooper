package effecthandler

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

// MoveCharacterHandler processes MoveCharacter effects.
type MoveCharacterHandler struct{}

func (h *MoveCharacterHandler) ResolveChoices(ge GameEngine, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	moveCharEffect := effect.GetMoveCharacter()
	if moveCharEffect == nil {
		return nil, fmt.Errorf("effect is not of type MoveCharacter")
	}
	return CreateChoicesFromSelector(ge, moveCharEffect.Target, payload, "Select character to move")
}

func (h *MoveCharacterHandler) Apply(ge GameEngine, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	moveCharEffect := effect.GetMoveCharacter()
	if moveCharEffect == nil {
		return fmt.Errorf("effect is not of type MoveCharacter")
	}

	state := ge.GetGameState()
	targetIDs, err := ge.ResolveSelectorToCharacters(state, moveCharEffect.Target, nil, payload, ability)
	if err != nil {
		return err
	}

	for _, targetID := range targetIDs {
		char := ge.GetCharacterByID(targetID)
		if char == nil {
			continue
		}
		// A generic move, let the moveCharacter logic handle the details.
		// The destination is specified in the effect, but the current moveCharacter implementation
		// in the engine doesn't use it. This could be improved.
		ge.MoveCharacter(char, 0, 0)
	}
	return nil
}

func (h *MoveCharacterHandler) GetDescription(effect *model.Effect) string {
	moveChar := effect.GetMoveCharacter()
	if moveChar == nil {
		return "(Invalid MoveCharacter effect)"
	}
	return fmt.Sprintf("Move character to %s", moveChar.Destination)
}
