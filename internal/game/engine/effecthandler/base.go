package effecthandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// GameEngine 提供了处理程序与游戏状态和引擎逻辑交互所需的方法。
// 此接口有助于将处理程序与主引擎包解耦。
type GameEngine interface {
	GetGameState() *model.GameState
	TriggerEvent(eventType model.GameEventType, payload *model.EventPayload)
	ResolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, ctx *EffectContext) ([]int32, error)
	GetCharacterByID(id int32) *model.Character
	MoveCharacter(char *model.Character, dx, dy int)
}

// EffectContext 为效果解析和应用提供上下文信息。
type EffectContext struct {
	Ability *model.Ability
	Payload *model.UseAbilityPayload
	Choice  *model.ChooseOptionPayload
}

// EffectHandler 定义了处理特定类型游戏效果的接口。
type EffectHandler interface {
	// ResolveChoices 检查效果是否需要玩家选择，并返回可用选项。
	ResolveChoices(ge GameEngine, effect *model.Effect, ctx *EffectContext) ([]*model.Choice, error)

	// Apply 执行效果的逻辑，应用状态更改并发布事件。
	Apply(ge GameEngine, effect *model.Effect, ctx *EffectContext) error

	// GetDescription 返回效果的人类可读描述。
	GetDescription(effect *model.Effect) string
}
