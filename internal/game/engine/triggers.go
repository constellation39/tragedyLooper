package engine

import (
	model "tragedylooper/internal/game/proto/v1"

	"go.uber.org/zap"
)

// checkAndTriggerAbilities 遍历所有角色的能力，并触发与给定触发器匹配的能力。
// event 参数是可选的，仅在 triggerType 为 ON_GAME_EVENT 时使用。
func (ge *GameEngine) checkAndTriggerAbilities(triggerType model.TriggerType, event *model.GameEvent) {
	ge.logger.Debug("Checking for abilities to trigger", zap.String("triggerType", triggerType.String()))

	for _, char := range ge.GameState.Characters {
		for i, ability := range char.Abilities {
			if ability.TriggerType != triggerType {
				continue
			}

			// 如果是事件驱动的，请检查事件过滤器
			if triggerType == model.TriggerType_ON_GAME_EVENT {
				if event == nil || !ge.eventMatchesFilter(event, ability.EventFilters) {
					continue
				}
			}

			// 检查是否已经使用过（如果适用）
			if ability.OncePerLoop && ability.UsedThisLoop {
				continue
			}

			// TODO: 检查其他条件（例如，目标有效性）

			ge.logger.Info("Triggering ability", zap.String("character", char.Name), zap.String("ability", ability.Name))

			// 简单的自动效果应用
			// 对于需要玩家选择的目标，这将需要一个更复杂的流程
			payload := &model.UseAbilityPayload{CharacterId: char.Id, AbilityId: ability.Id} // 假设自我目标
			if err := ge.applyEffect(ability.Effect, ability, payload); err != nil {
				ge.logger.Error("Error applying triggered ability effect", zap.Error(err))
			}

			if ability.OncePerLoop {
				ge.GameState.Characters[char.Id].Abilities[i].UsedThisLoop = true
			}

						// ge.publishGameEvent(model.GameEventType_ABILITY_USED, &model.AbilityUsedEvent{CharacterId: char.Id, AbilityName: ability.Name})
		}
	}
}

// eventMatchesFilter 检查给定事件的类型是否在过滤列表中。
func (ge *GameEngine) eventMatchesFilter(event *model.GameEvent, filters []model.GameEventType) bool {
	if len(filters) == 0 {
		return true // 空过滤器匹配任何事件
	}
	for _, filter := range filters {
		if event.Type == filter {
			return true
		}
	}
	return false
}
