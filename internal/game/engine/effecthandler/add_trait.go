package effecthandler // 定义效果处理器的包

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

// init 函数在包加载时自动执行，用于注册 AddTrait 效果处理器。
func init() {
	Register[*model.Effect_AddTrait](&AddTraitHandler{})
}

// AddTraitHandler 结构体实现了处理 AddTrait 效果的逻辑。
// AddTrait 效果用于给指定角色添加一个特性（Trait）。
type AddTraitHandler struct{}

func (h *AddTraitHandler) ResolveChoices(ge GameEngine, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	addTraitEffect := effect.GetAddTrait()
	if addTraitEffect == nil {
		return nil, fmt.Errorf("effect is not of type AddTrait")
	}
	// 根据效果的目标选择器创建选项，让玩家选择要添加特性的角色。
	return CreateChoicesFromSelector(ge, addTraitEffect.Target, payload, "Select character to add trait to")
}

func (h *AddTraitHandler) Apply(ge GameEngine, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	addTraitEffect := effect.GetAddTrait()
	if addTraitEffect == nil {
		return fmt.Errorf("effect is not of type AddTrait")
	}

	state := ge.GetGameState()
	// 解析目标选择器，获取所有受影响的角色ID。
	targetIDs, err := ge.ResolveSelectorToCharacters(state, addTraitEffect.Target, nil, payload, ability)
	if err != nil {
		return err
	}

	// 遍历所有目标角色，为每个角色添加特性并发布 TraitAdded 事件。
	for _, targetID := range targetIDs {
		event := &model.TraitAddedEvent{CharacterId: targetID, Trait: addTraitEffect.Trait}
		ge.ApplyAndPublishEvent(model.GameEventType_TRAIT_ADDED, event)
	}
	return nil
}

func (h *AddTraitHandler) GetDescription(effect *model.Effect) string {
	addTrait := effect.GetAddTrait()
	if addTrait == nil {
		return "(无效的 AddTrait 效果)"
	}
	// 返回 AddTrait 效果的描述字符串。
	return fmt.Sprintf("添加特征 '%s'", addTrait.Trait)
}