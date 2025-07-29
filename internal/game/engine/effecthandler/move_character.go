package effecthandler

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register[*model.Effect_MoveCharacter](&MoveCharacterHandler{})
}

// MoveCharacterHandler 处理 MoveCharacter 效果。
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
		// 通用移动，让 moveCharacter 逻辑处理细节。
		// 目的地在效果中指定，但引擎中当前的 moveCharacter 实现
		// 并未使用它。这可以改进。
		ge.MoveCharacter(char, 0, 0)
	}
	return nil
}

func (h *MoveCharacterHandler) GetDescription(effect *model.Effect) string {
	moveChar := effect.GetMoveCharacter()
	if moveChar == nil {
		return "(无效的 MoveCharacter 效果)"
	}
	return fmt.Sprintf("将角色移动到 %s", moveChar.Destination)
}