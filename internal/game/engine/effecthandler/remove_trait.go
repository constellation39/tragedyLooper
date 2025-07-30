package effecthandler

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

// init 函数在包加载时自动执行，用于注册 RemoveTrait 效果处理器。
func init() {
	Register[*model.Effect_RemoveTrait](&RemoveTraitHandler{})
}

// RemoveTraitHandler 结构体实现了处理 RemoveTrait 效果的逻辑。
// RemoveTrait 效果用于从指定角色移除一个特性（Trait）。
type RemoveTraitHandler struct{}

func (h *RemoveTraitHandler) ResolveChoices(ge GameEngine, effect *model.Effect, ctx *EffectContext) ([]*model.Choice, error) {
	removeTraitEffect := effect.GetRemoveTrait()
	if removeTraitEffect == nil {
		return nil, fmt.Errorf("effect is not of type RemoveTrait")
	}
	// 根据效果的目标选择器创建选项，让玩家选择要移除特性的角色。
	return CreateChoicesFromSelector(ge, removeTraitEffect.Target, ctx, "Select character to remove trait from")
}

func (h *RemoveTraitHandler) Apply(ge GameEngine, effect *model.Effect, ctx *EffectContext) error {
	removeTraitEffect := effect.GetRemoveTrait()
	if removeTraitEffect == nil {
		return fmt.Errorf("effect is not of type RemoveTrait")
	}

	state := ge.GetGameState()
	// 解析目标选择器，获取所有受影响的角色ID。
	targetIDs, err := ge.ResolveSelectorToCharacters(state, removeTraitEffect.Target, ctx)
	if err != nil {
		return err
	}

	// 遍历所有目标角色，为每个角色移除特性并发布 TraitRemovedEvent 事件。
	for _, targetID := range targetIDs {
		event := &model.TraitRemovedEvent{CharacterId: targetID, Trait: removeTraitEffect.Trait}
		ge.ApplyAndPublishEvent(model.GameEventType_TRAIT_REMOVED, &model.EventPayload{
			Payload: &model.EventPayload_TraitRemoved{TraitRemoved: event},
		})
	}
	return nil
}

func (h *RemoveTraitHandler) GetDescription(effect *model.Effect) string {
	removeTrait := effect.GetRemoveTrait()
	if removeTrait == nil {
		return "(无效的 RemoveTrait 效果)"
	}
	// 返回 RemoveTrait 效果的描述字符串。
	return fmt.Sprintf("移除特征 '%s'", removeTrait.Trait)
}