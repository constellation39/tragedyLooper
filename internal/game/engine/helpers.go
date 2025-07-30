package engine // 定义游戏引擎包

import (
	model "tragedylooper/pkg/proto/v1"
)

// getProtagonistPlayers 获取所有主角玩家。
// 返回值: 主角玩家列表。
func (ge *GameEngine) getProtagonistPlayers() []*model.Player {
	var protagonists []*model.Player
	for _, p := range ge.GameState.Players {
		if p.Role == model.PlayerRole_PROTAGONIST {
			protagonists = append(protagonists, p)
		}
	}
	return protagonists
}

// getMastermindPlayer 获取主谋玩家。
// 返回值: 主谋玩家对象，如果不存在则返回 nil。
func (ge *GameEngine) getMastermindPlayer() *model.Player {
	for _, p := range ge.GameState.Players {
		if p.Role == model.PlayerRole_MASTERMIND {
			return p
		}
	}
	return nil
}

// isMastermindReady 检查主谋是否已准备好。
// 返回值: 如果主谋已准备好或不存在主谋，则返回 true；否则返回 false。
func (ge *GameEngine) isMastermindReady() bool {
	mm := ge.getMastermindPlayer()
	if mm == nil {
		return true // 没有主谋，所以他们自然就准备好了
	}
	return ge.playerReady[mm.Id]
}

// areAllPlayersReady 检查所有玩家是否都已准备好。
// 返回值: 如果所有玩家都已准备好，则返回 true；否则返回 false。
func (ge *GameEngine) areAllPlayersReady() bool {
	for _, p := range ge.GameState.Players {
		if !ge.playerReady[p.Id] {
			return false
		}
	}
	return true
}

// checkConditions 检查给定条件是否满足。
// conditions: 要检查的条件列表。
// player: 相关的玩家（如果适用）。
// payload: 相关的操作负载（如果适用）。
// ability: 相关的能力（如果适用）。
// 返回值: 如果所有条件都满足，则返回 true；否则返回 false。
func (ge *GameEngine) checkConditions(conditions []*model.Condition, player *model.Player, payload *model.UseAbilityPayload, ability *model.Ability) bool {
	return true
}

// checkGameEndConditions 检查游戏结束条件是否满足。
// 返回值: 一个布尔值，表示游戏是否结束，以及获胜方的角色类型。
func (ge *GameEngine) checkGameEndConditions() (bool, model.PlayerRole) {
	return false, model.PlayerRole_PLAYER_ROLE_UNSPECIFIED
}
