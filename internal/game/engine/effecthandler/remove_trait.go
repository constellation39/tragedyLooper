package effecthandler

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register[*model.Effect_RemoveTrait](&RemoveTraitHandler{})
}

// RemoveTraitHandler processes RemoveTrait effects.
type RemoveTraitHandler struct{}

func (h *RemoveTraitHandler) ResolveChoices(ge GameEngine, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	removeTraitEffect := effect.GetRemoveTrait()
	if removeTraitEffect == nil {
		return nil, fmt.Errorf("effect is not of type RemoveTrait")
	}
	return CreateChoicesFromSelector(ge, removeTraitEffect.Target, payload, "Select character to remove trait from")
}

func (h *RemoveTraitHandler) Apply(ge GameEngine, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	removeTraitEffect := effect.GetRemoveTrait()
	if removeTraitEffect == nil {
		return fmt.Errorf("effect is not of type RemoveTrait")
	}

	state := ge.GetGameState()
	targetIDs, err := ge.ResolveSelectorToCharacters(state, removeTraitEffect.Target, nil, payload, ability)
	if err != nil {
		return err
	}

	for _, targetID := range targetIDs {
		event := &model.TraitRemovedEvent{CharacterId: targetID, Trait: removeTraitEffect.Trait}
		ge.ApplyAndPublishEvent(model.GameEventType_TRAIT_REMOVED, event)
	}
	return nil
}

func (h *RemoveTraitHandler) GetDescription(effect *model.Effect) string {
	removeTrait := effect.GetRemoveTrait()
	if removeTrait == nil {
		return "(Invalid RemoveTrait effect)"
	}
	return fmt.Sprintf("Remove trait '%s'", removeTrait.Trait)
}
