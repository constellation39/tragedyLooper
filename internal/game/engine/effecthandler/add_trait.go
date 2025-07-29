package effecthandler // Defines the package for effect handlers

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

// init automatically executes when the package is loaded, registering the AddTrait effect handler.
func init() {
	Register[*model.Effect_AddTrait](&AddTraitHandler{})
}

// AddTraitHandler implements the logic for handling the AddTrait effect.
// The AddTrait effect is used to add a trait to a specified character.
type AddTraitHandler struct{}

func (h *AddTraitHandler) ResolveChoices(ge GameEngine, effect *model.Effect, payload *model.UseAbilityPayload) ([]*model.Choice, error) {
	addTraitEffect := effect.GetAddTrait()
	if addTraitEffect == nil {
		return nil, fmt.Errorf("effect is not of type AddTrait")
	}
	// Create choices from the effect's target selector, allowing the player to choose which character to add the trait to.
	return CreateChoicesFromSelector(ge, addTraitEffect.Target, payload, "Select character to add trait to")
}

func (h *AddTraitHandler) Apply(ge GameEngine, effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	addTraitEffect := effect.GetAddTrait()
	if addTraitEffect == nil {
		return fmt.Errorf("effect is not of type AddTrait")
	}

	state := ge.GetGameState()
	// Resolve the target selector to get all affected character IDs.
	targetIDs, err := ge.ResolveSelectorToCharacters(state, addTraitEffect.Target, nil, payload, ability)
	if err != nil {
		return err
	}

	// Iterate over all target characters, add the trait to each, and publish a TraitAdded event.
	for _, targetID := range targetIDs {
		event := &model.TraitAddedEvent{CharacterId: targetID, Trait: addTraitEffect.Trait}
		ge.ApplyAndPublishEvent(model.GameEventType_TRAIT_ADDED, event)
	}
	return nil
}

func (h *AddTraitHandler) GetDescription(effect *model.Effect) string {
	addTrait := effect.GetAddTrait()
	if addTrait == nil {
		return "(Invalid AddTrait effect)"
	}
	// Returns the description string for the AddTrait effect.
	return fmt.Sprintf("Add trait '%s'", addTrait.Trait)
}