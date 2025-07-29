package engine

import (
	model "tragedylooper/internal/game/proto/v1"

	"go.uber.org/zap"
)

// checkAndTriggerAbilities iterates through all character abilities and triggers those that match the given trigger type.
func (ge *GameEngine) checkAndTriggerAbilities(triggerType model.TriggerType) {
	ge.logger.Debug("Checking for abilities to trigger", zap.String("triggerType", triggerType.String()))
	if true {
		return
	}
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
			if err := ge.applyEffect(ability.Config.Effect, ability, nil, nil); err != nil {
				ge.logger.Error("Error applying triggered ability effect", zap.Error(err))
			}

			if ability.Config.OncePerLoop {
				ability.UsedThisLoop = true // Mark the instance of the ability on the character as used
			}
		}
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
