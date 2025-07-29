package engine // 定义游戏引擎包

import (
	"fmt"
	"tragedylooper/internal/game/engine/effecthandler" // 导入效果处理程序包
	model "tragedylooper/pkg/proto/v1"                 // 导入协议缓冲区模型
)

// applyEffect 查找效果的适当处理程序，并使用它来解决选择，然后应用效果。
// effect: 要应用的效果。
// ability: 触发此效果的能力（如果适用）。
// payload: 玩家操作的有效负载（如果适用）。
// choice: 玩家做出的选择（如果适用）。
func (ge *GameEngine) applyEffect(effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	handler, err := effecthandler.GetEffectHandler(effect)
	if err != nil {
		return err
	}

	// 1. 解决选择
	// 如果需要选择，则发布一个事件并等待玩家输入。
	// 效果的实际应用将在做出选择后发生。
	choices, err := handler.ResolveChoices(ge, effect, payload)
	if err != nil {
		return fmt.Errorf("error resolving choices: %w", err)
	}

	if len(choices) > 0 && choice == nil {
		choiceEvent := &model.ChoiceRequiredEvent{Choices: choices}
		ge.ApplyAndPublishEvent(model.GameEventType_CHOICE_REQUIRED, &model.EventPayload{
			Payload: &model.EventPayload_ChoiceRequired{ChoiceRequired: choiceEvent},
		})
		return nil // 停止处理，直到收到选择
	}

	// 2. 应用效果
	// 如果不需要选择，或者已经提供了选择，则应用效果。
	err = handler.Apply(ge, effect, ability, payload, choice)
	if err != nil {
		return fmt.Errorf("error applying effect: %w", err)
	}

	return nil
}

// GetEffectDescription 查找适当的处理程序并返回效果的描述。
// effect: 要获取描述的效果。
// 返回值: 效果的描述字符串。
func (ge *GameEngine) GetEffectDescription(effect *model.Effect) string {
	return effecthandler.GetEffectDescription(ge, effect)
}
