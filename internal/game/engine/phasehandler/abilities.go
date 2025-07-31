package phasehandler

import (
	"time"
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

type AbilityTurn int

const (
	MastermindAbilityTurn AbilityTurn = iota
	ProtagonistAbilityTurn
)

// AbilitiesPhase 是玩家可以使用角色能力的阶段。
type AbilitiesPhase struct {
	basePhase
	turn                 AbilityTurn
	protagonistTurnIndex int
}

// Type 返回阶段类型。
func (p *AbilitiesPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_ABILITIES }

// Enter 在阶段开始时调用。
func (p *AbilitiesPhase) Enter(ge GameEngine) Phase {
	p.turn = MastermindAbilityTurn
	p.protagonistTurnIndex = 0
	ge.ResetPlayerReadiness()

	// 可选：为主谋触发 AI
	// ge.TriggerAIPlayerAction(ge.GetMastermindPlayer().Id)

	return nil
}

// HandleAction 处理来自玩家的行动。
func (p *AbilitiesPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if !p.isActionInTurn(ge, player) {
		ge.Logger().Warn("Received action from player out of turn", zap.String("player", player.Name))
		return nil
	}

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_UseAbility:
		p.handleUseAbilityAction(ge, player, payload.UseAbility)
		// 注意：使用能力不会自动结束回合。
		// 玩家必须明确跳过。
	case *model.PlayerActionPayload_PassTurn:
		return p.handlePassTurn(ge, player)
	}

	return nil
}

// HandleTimeout 处理超时。
func (p *AbilitiesPhase) HandleTimeout(ge GameEngine) Phase {
	ge.Logger().Info("Abilities phasehandler timed out, passing turn.")
	var player *model.Player
	if p.turn == MastermindAbilityTurn {
		player = ge.GetMastermindPlayer()
	} else {
		protagonists := ge.GetProtagonistPlayers()
		if p.protagonistTurnIndex < len(protagonists) {
			player = protagonists[p.protagonistTurnIndex]
		}
	}
	if player != nil {
		return p.handlePassTurn(ge, player)
	}
	return &IncidentsPhase{}
}

// TimeoutDuration 返回此阶段的超时持续时间。
func (p *AbilitiesPhase) TimeoutDuration() time.Duration { return 60 * time.Second }

func (p *AbilitiesPhase) isActionInTurn(ge GameEngine, player *model.Player) bool {
	if p.turn == MastermindAbilityTurn {
		return player.Role == model.PlayerRole_PLAYER_ROLE_MASTERMIND
	}

	protagonists := ge.GetProtagonistPlayers()
	if p.protagonistTurnIndex >= len(protagonists) {
		return false // 不应该发生
	}
	return player.Id == protagonists[p.protagonistTurnIndex].Id
}

func (p *AbilitiesPhase) handlePassTurn(ge GameEngine, player *model.Player) Phase {
	ge.Logger().Info("Player passed ability turn", zap.String("player", player.Name))

	if p.turn == MastermindAbilityTurn {
		p.turn = ProtagonistAbilityTurn
		ge.Logger().Info("Transitioning to Protagonist ability turn")
		// 可选：为第一个主角触发 AI
		// protagonists := ge.GetProtagonistPlayers()
		// if len(protagonists) > 0 {
		// 	ge.TriggerAIPlayerAction(protagonists[0].Id)
		// }
		return nil
	}

	p.protagonistTurnIndex++
	protagonists := ge.GetProtagonistPlayers()
	if p.protagonistTurnIndex >= len(protagonists) {
		ge.Logger().Info("All protagonists have acted, moving to Incidents Phase")
		return &IncidentsPhase{}
	}

	// 可选：为下一个主角触发 AI
	// ge.TriggerAIPlayerAction(protagonists[p.protagonistTurnIndex].Id)
	return nil
}

func (p *AbilitiesPhase) handleUseAbilityAction(ge GameEngine, player *model.Player, payload *model.UseAbilityPayload) {
	char, ok := ge.GetGameState().Characters[payload.CharacterId]
	if !ok {
		ge.Logger().Warn("Character not found for ability use", zap.Int32("characterID", payload.CharacterId))
		return
	}

	var ability *model.Ability
	abilityFound := false
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

	if ability.UsedThisLoop {
		ge.Logger().Warn("Ability has already been used this loop", zap.String("abilityName", ability.Config.Name))
		return
	}

	// 在这里，我们需要检查玩家是否有权使用此能力。
	// 目前，我们假设如果轮到他们，他们就可以。
	// 这里可能需要对好感能力进行更复杂的检查。

	if err := ge.ApplyEffect(ability.Config.Effect, ability, payload, nil); err != nil {
		ge.Logger().Error("Failed to apply effect for ability", zap.String("abilityName", ability.Config.Name), zap.Error(err))
		return
	}

	if ability.Config.OncePerLoop {
		ability.UsedThisLoop = true
	}

	ge.Logger().Info("Player used ability", zap.String("player", player.Name), zap.String("ability", ability.Config.Name))
}
