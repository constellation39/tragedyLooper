package engine // 定义游戏引擎包

import (
	"time"
	"tragedylooper/internal/game/loader" // 导入游戏数据加载器
	model "tragedylooper/pkg/proto/v1"   // 导入协议缓冲区模型

	"go.uber.org/zap" // 导入 Zap 日志库
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

// initializeGameStateFromScript 根据剧本初始化游戏状态。
// gameConfig: 游戏配置。
// playerMap: 玩家ID到玩家对象的映射。
func (ge *GameEngine) initializeGameStateFromScript(playerMap map[int32]*model.Player) {
	characters := make(map[int32]*model.Character)
	for _, charInScript := range ge.gameConfig.GetScript().Characters {

		charData, err := loader.Get[*model.CharacterConfig](ge.gameConfig, charInScript.CharacterId)
		if err != nil {
			ge.logger.Warn("character config not found", zap.Int32("characterID", charInScript.CharacterId))
			continue
		}

		abilities := make([]*model.Ability, 0)

		for _, ab := range charData.AbilityIds {
			ability, err := loader.Get[*model.AbilityConfig](ge.gameConfig, ab)
			if err != nil {
				ge.logger.Warn("ability config not found", zap.Int32("abilityID", ab))
				continue
			}
			abilities = append(abilities, &model.Ability{
				Config:           ability,
				UsedThisLoop:     false,
				OwnerCharacterId: 0,
			})
		}

		characters[charInScript.CharacterId] = &model.Character{
			Config:          charData,
			CurrentLocation: charInScript.InitialLocation,
			Paranoia:        charInScript.InitialParanoia,
			Goodwill:        charInScript.InitialGoodwill,
			Intrigue:        charInScript.InitialIntrigue,
			HiddenRole:      charInScript.HiddenRole,
			Abilities:       abilities,
			IsAlive:         true,
			InPanicMode:     false,
			Traits:          charData.Traits,
		}
	}

	ge.GameState = &model.GameState{
		GameId:                  "new_game", // 应该生成
		Characters:              characters,
		Players:                 playerMap,
		CurrentDay:              1,
		CurrentLoop:             1,
		CurrentPhase:            ge.pm.CurrentPhase().Type(),
		ActiveTragedies:         make(map[int32]bool),
		PreventedTragedies:      make(map[int32]bool),
		PlayedCardsThisDay:      make(map[int32]*model.CardList),
		PlayedCardsThisLoop:     make(map[int32]bool),
		LastUpdateTime:          time.Now().Unix(),
		DayEvents:               make([]*model.GameEvent, 0),
		LoopEvents:              make([]*model.GameEvent, 0),
		CharacterParanoiaLimits: make(map[int32]int32),
		CharacterGoodwillLimits: make(map[int32]int32),
		CharacterIntrigueLimits: make(map[int32]int32),
	}
}

// dealInitialCards 根据剧本为玩家分发初始卡牌。
func (ge *GameEngine) dealInitialCards() {
	script := ge.gameConfig.GetScript()
	if script == nil {
		ge.logger.Error("cannot deal cards, script not loaded")
		return
	}

	mastermind := ge.getMastermindPlayer()
	if mastermind != nil {
		for _, cardID := range script.MastermindCardIds {
			cardConfig, err := loader.Get[*model.CardConfig](ge.gameConfig, cardID)
			if err != nil {
				ge.logger.Warn("mastermind card config not found", zap.Int32("cardID", cardID))
				continue
			}
			mastermind.Hand = append(mastermind.Hand, &model.Card{Config: cardConfig})
		}
	}

	protagonists := ge.getProtagonistPlayers()
	for _, protagonist := range protagonists {
		for _, cardID := range script.ProtagonistCardIds {
			cardConfig, err := loader.Get[*model.CardConfig](ge.gameConfig, cardID)
			if err != nil {
				ge.logger.Warn("protagonist card config not found", zap.Int32("cardID", cardID))
				continue
			}
			protagonist.Hand = append(protagonist.Hand, &model.Card{Config: cardConfig})
		}
	}
}
