package effecthandler // 定义效果处理器的包

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

// init 函数在包加载时自动执行，用于注册 MoveCharacter 效果处理器。
func init() {
	Register[*model.Effect_MoveCharacter](&MoveCharacterHandler{})
}

// MoveCharacterHandler 结构体实现了处理 MoveCharacter 效果的逻辑。
// MoveCharacter 效果用于移动指定角色到特定位置。
type MoveCharacterHandler struct{}

func (h *MoveCharacterHandler) ResolveChoices(ge GameEngine, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	moveCharEffect := effect.GetMoveCharacter()
	if moveCharEffect == nil {
		return nil, fmt.Errorf("effect is not of type MoveCharacter")
	}
	// 根据效果的目标选择器创建选项，让玩家选择要移动的角色。
	return CreateChoicesFromSelector(ge, moveCharEffect.Target, payload, "Select character to move")
}

func (h *MoveCharacterHandler) Apply(ge GameEngine, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	moveCharEffect := effect.GetMoveCharacter()
	if moveCharEffect == nil {
		return fmt.Errorf("effect is not of type MoveCharacter")
	}

	state := ge.GetGameState()
	// 解析目标选择器，获取所有受影响的角色ID。
	targetIDs, err := ge.ResolveSelectorToCharacters(state, moveCharEffect.Target, nil, payload, ability)
	if err != nil {
		return err
	}

	// 遍历所有目标角色，并调用 GameEngine 的 MoveCharacter 方法移动角色。
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
	// 返回 MoveCharacter 效果的描述字符串。
	return fmt.Sprintf("将角色移动到 %s", moveChar.Destination)
}
