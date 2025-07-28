package engine

import (
	"errors"
	model "tragedylooper/internal/game/proto/v1"

	"go.uber.org/zap"
)

// checkAndTriggerAbilities iterates through all character abilities and triggers those that match the given trigger type.
func (ge *GameEngine) checkAndTriggerAbilities(triggerType model.TriggerType) {
	ge.logger.Debug("Checking for abilities to trigger", zap.String("triggerType", triggerType.String()))

	for _, char := range ge.GameState.Characters {
		for _, ability := range char.Abilities {
			if ability.Config.TriggerType != triggerType {
				continue
			}

			if ability.Config.OncePerLoop && ability.UsedThisLoop {
				continue
			}

			// For abilities that trigger automatically, we pass nil for player and payload initially.
			if !ge.checkConditions(ability.Config.Conditions, nil, nil, ability) {
				continue
			}

			// If the ability requires a choice, we need to ask the player.
			if ability.Config.RequiresChoice {
				// TODO: Implement logic to send a ChoiceRequiredEvent to the player who controls the character.
				ge.logger.Info("Ability requires choice, skipping automatic trigger for now.", zap.String("ability", ability.Config.Name))
				continue
			}

			ge.logger.Info("Auto-triggering ability", zap.String("character", char.Config.Name), zap.String("ability", ability.Config.Name))

			// Since this is an automatic trigger, we assume no specific player action payload.
			if err := ge.applyEffect(ability.Config.Effect, nil, nil, ability); err != nil {
				ge.logger.Error("Error applying triggered ability effect", zap.Error(err))
			}

			if ability.Config.OncePerLoop {
				ability.UsedThisLoop = true // Mark the instance of the ability on the character as used
			}
		}
	}
}

// resolveSelectorToCharacters determines the character IDs targeted by a selector.
func (ge *GameEngine) resolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, player *model.Player, payload *model.UseAbilityPayload, ability *model.Ability) ([]int32, error) {
	if sel == nil {
		return nil, errors.New("target selector is nil")
	}

	switch sel.Type {
	case model.TargetSelector_ABILITY_USER:
		if ability == nil {
			return nil, errors.New("ability is nil for ABILITY_USER selector")
		}
		return []int32{ability.OwnerCharacterId}, nil
	case model.TargetSelector_ABILITY_TARGET:
		if payload == nil {
			return nil, errors.New("payload is nil for ABILITY_TARGET selector")
		}
		if targetChar, ok := payload.Target.(*model.UseAbilityPayload_TargetCharacterId); ok {
			return []int32{targetChar.TargetCharacterId}, nil
		}
		return nil, errors.New("payload does not contain a target character for ABILITY_TARGET selector")
	case model.TargetSelector_ALL_CHARACTERS:
		ids := make([]int32, 0, len(gs.Characters))
		for id := range gs.Characters {
			ids = append(ids, id)
		}
		return ids, nil
	case model.TargetSelector_SPECIFIC_CHARACTER:
		return []int32{sel.GetSpecificCharacterId()}, nil
	// TODO: Implement other selector types
	default:
		return nil, errors.New("unsupported target selector type")
	}
}

// cloneAbility creates a deep copy of an ability to avoid modifying the original config.
func cloneAbility(original *model.Ability) *model.Ability {
	if original == nil {
		return nil
	}
	// A proper deep copy would be needed here. Using proto.Clone for simplicity.
	// return proto.Clone(original).(*model.Ability)
	// Manual clone for now:
	return &model.Ability{
		Config:           original.Config, // This should be a deep copy if mutable
		OwnerCharacterId: original.OwnerCharacterId,
		UsedThisLoop:     original.UsedThisLoop,
	}
}