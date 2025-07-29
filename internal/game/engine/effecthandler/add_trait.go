package effecthandler

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register[*model.Effect_AddTrait](&AddTraitHandler{})
}

// AddTraitHandler 处理 AddTrait 效果。
type AddTraitHandler struct{}

func (h *AddTraitHandler) ResolveChoices(ge GameEngine, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	addTraitEffect := effect.GetAddTrait()
	if addTraitEffect == nil {
		return nil, fmt.Errorf("effect is not of type AddTrait")
	}
	return CreateChoicesFromSelector(ge, addTraitEffect.Target, payload, "Select character to add trait to")
}

func (h *AddTraitHandler) Apply(ge GameEngine, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	addTraitEffect := effect.GetAddTrait()
	if addTraitEffect == nil {
		return fmt.Errorf("effect is not of type AddTrait")
	}

	state := ge.GetGameState()
	targetIDs, err := ge.ResolveSelectorToCharacters(state, addTraitEffect.Target, nil, payload, ability)
	if err != nil {
		return err
	}

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
	return fmt.Sprintf("添加特征 '%s'", addTrait.Trait)
}