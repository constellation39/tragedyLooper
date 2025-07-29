package engine

import (
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// checkAndTriggerAbilities 遍历所有角色能力并触发与给定触发器类型匹配的能力。
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

			// 对于自动触发的能力，我们最初为玩家和有效负载传递 nil。
			if !ge.checkConditions(ability.Config.Conditions, nil, nil, ability) {
				continue
			}

			// 如果能力需要选择，我们需要询问玩家。
			if ability.Config.RequiresChoice {
				// TODO: 实现向控制角色的玩家发送 ChoiceRequiredEvent 的逻辑。
				ge.logger.Info("Ability requires choice, skipping automatic trigger for now.", zap.String("ability", ability.Config.Name))
				continue
			}

			ge.logger.Info("Auto-triggering ability", zap.String("character", char.Config.Name), zap.String("ability", ability.Config.Name))

			// 由于这是一个自动触发器，我们假设没有特定的玩家操作有效负载。
			if err := ge.applyEffect(ability.Config.Effect, ability, nil, nil); err != nil {
				ge.logger.Error("Error applying triggered ability effect", zap.Error(err))
			}

			if ability.Config.OncePerLoop {
				ability.UsedThisLoop = true // 将角色身上的能力实例标记为已使用
			}
		}
	}
}

// cloneAbility 创建能力的深层副本，以避免修改原始配置。
func cloneAbility(original *model.Ability) *model.Ability {
	if original == nil {
		return nil
	}
	// 这里需要一个正确的深层副本。为简单起见，使用 proto.Clone。
	// return proto.Clone(original).(*model.Ability)
	// 目前手动克隆：
	return &model.Ability{
		Config:           original.Config, // 如果可变，则应为深层副本
		OwnerCharacterId: original.OwnerCharacterId,
		UsedThisLoop:     original.UsedThisLoop,
	}
}