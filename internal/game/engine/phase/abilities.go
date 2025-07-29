package phase // 定义游戏阶段包

import (
	model "tragedylooper/pkg/proto/v1" // 导入协议缓冲区模型

	"go.uber.org/zap" // 导入 Zap 日志库
)

// AbilitiesPhase 能力阶段，玩家可以在此阶段使用角色能力。
type AbilitiesPhase struct{ basePhase }

// Type 返回阶段类型，表示当前是能力阶段。
func (p *AbilitiesPhase) Type() model.GamePhase { return model.GamePhase_ABILITIES }

// Enter 进入能力阶段。
// ge: 游戏引擎接口。
// 返回值: 下一个阶段的实例。
func (p *AbilitiesPhase) Enter(ge GameEngine) Phase {
	// 玩家可以使用能力。
	// 这个阶段可能需要玩家输入并有超时。
	return &IncidentsPhase{}
}

// HandleAction 处理玩家在能力阶段的操作。
// ge: 游戏引擎接口。
// playerID: 执行操作的玩家ID。
// action: 玩家操作的负载。
// 返回值: 如果阶段发生变化，则返回新的阶段实例；否则返回 nil。
func (p *AbilitiesPhase) HandleAction(ge GameEngine, playerID int32, action *model.PlayerActionPayload) Phase {
	state := ge.GetGameState()
	player, ok := state.Players[playerID]
	if !ok {
		ge.Logger().Warn("Action from unknown player", zap.Int32("playerID", playerID))
		return nil
	}

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_UseAbility:
		handleUseAbilityAction(ge, player, payload.UseAbility)
	}
	return nil
}

func handleUseAbilityAction(ge GameEngine, player *model.Player, payload *model.UseAbilityPayload) {
	var ability *model.Ability
	abilityFound := false
	char, ok := ge.GetGameState().Characters[payload.CharacterId]
	if !ok {
		ge.Logger().Warn("Character not found for ability use", zap.Int32("characterID", payload.CharacterId))
		return
	}

	for i := range char.Abilities {
		if char.Abilities[i].Config.Id == payload.AbilityId {
			ability = char.Abilities[i]
			abilityFound = true
			break
		}
	}

	if !abilityFound {
		ge.Logger().Warn("Ability not found on character", zap.Int32("abilityID", payload.AbilityId), zap.Int32("characterID", payload.CharacterId))
		return
	}

	// TODO: 我们需要重新实现 applyEffect，因为它不是 GameEngine 接口的一部分。
	// if err := ge.applyEffect(ability.Config.Effect, ability, payload, nil); err != nil {
	// 	ge.Logger().Error("Failed to apply effect for ability", zap.String("abilityName", ability.Config.Name), zap.Error(err))
	// 	return
	// }

	// 如果能力配置为每回合只能使用一次，则标记为已使用。
	if ability.Config.OncePerLoop {
		ability.UsedThisLoop = true
	}
	// 注意：使用能力不会自动使玩家“准备好”。
	// 他们必须使用 PassTurnAction 明确地跳过他们的回合。
}