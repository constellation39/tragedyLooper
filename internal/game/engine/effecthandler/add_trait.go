package effecthandler

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

// init 函数在包加载时自动执行，注册 AddTrait 效果处理器。
func init() {
	Register[*model.Effect_AddTrait](&AddTraitHandler{})
}

// AddTraitHandler 实现处理 AddTrait 效果的逻辑。
// AddTrait 效果用于向指定角色添加特征。
type AddTraitHandler struct{}

func (h *AddTraitHandler) ResolveChoices(ge GameEngine, effect *model.Effect, ctx *EffectContext) ([]*model.Choice, error) {
	addTraitEffect := effect.GetAddTrait()
	if addTraitEffect == nil {
		return nil, fmt.Errorf("effect is not of type AddTrait")
	}
	// 根据效果的目标选择器创建选项，让玩家选择要向哪个角色添加特征。
	return CreateChoicesFromSelector(ge, addTraitEffect.Target, ctx, "Select character to add trait to")
}

func (h *AddTraitHandler) Apply(ge GameEngine, effect *model.Effect, ctx *EffectContext) error {
	addTraitEffect := effect.GetAddTrait()
	if addTraitEffect == nil {
		return fmt.Errorf("effect is not of type AddTrait")
	}

	state := ge.GetGameState()
	// 解析目标选择器以获取所有受影响的角色 ID。
	targetIDs, err := ge.ResolveSelectorToCharacters(state, addTraitEffect.Target, ctx)
	if err != nil {
		return err
	}

	// 遍历所有目标角色，为每个角色添加特征，并发布 TraitAdded 事件。
	for _, targetID := range targetIDs {
		event := &model.TraitAddedEvent{CharacterId: targetID, Trait: addTraitEffect.Trait}
		ge.ApplyAndPublishEvent(model.GameEventType_TRAIT_ADDED, &model.EventPayload{
			Payload: &model.EventPayload_TraitAdded{TraitAdded: event},
		})
	}
	return nil
}

func (h *AddTraitHandler) GetDescription(effect *model.Effect) string {
	addTrait := effect.GetAddTrait()
	if addTrait == nil {
		return "(Invalid AddTrait effect)"
	}
	// 返回 AddTrait 效果的描述字符串。
	return fmt.Sprintf("Add trait '%s'", addTrait.Trait)
}
