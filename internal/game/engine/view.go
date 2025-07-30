package engine

import (
	"google.golang.org/protobuf/proto"

	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// GeneratePlayerView 为特定玩家创建游戏状态的过滤视图。
// 此方法不是线程安全的，必须仅在 runGameLoop goroutine 中调用。
func (ge *GameEngine) GeneratePlayerView(playerID int32) *model.PlayerView {
	player := ge.GameState.Players[playerID]
	if player == nil {
		return &model.PlayerView{}
	}

	view := &model.PlayerView{
		GameId:             ge.GameState.GameId,
		CurrentDay:         ge.GameState.CurrentDay,
		CurrentLoop:        ge.GameState.CurrentLoop,
		CurrentPhase:       ge.GameState.CurrentPhase,
		ActiveTragedies:    ge.GameState.ActiveTragedies,
		PreventedTragedies: ge.GameState.PreventedTragedies,
		PublicEvents:       ge.GameState.DayEvents,
	}

	// 根据玩家角色过滤角色信息
	view.Characters = make(map[int32]*model.PlayerViewCharacter)
	for id, char := range ge.GameState.Characters {
		// 创建一个副本以避免竞争和对核心状态的意外修改。
		charCopy := proto.Clone(char).(*model.Character)
		playerViewChar := &model.PlayerViewCharacter{
			Id:              id,
			Name:            charCopy.Config.Name,
			Traits:          charCopy.Traits,
			CurrentLocation: charCopy.CurrentLocation,
			Paranoia:        charCopy.Paranoia,
			Goodwill:        charCopy.Goodwill,
			Intrigue:        charCopy.Intrigue,
			Abilities:       charCopy.Abilities,
			IsAlive:         charCopy.IsAlive,
			InPanicMode:     charCopy.InPanicMode,
			Rules:           charCopy.Config.Rules,
		}
		if player.Role == model.PlayerRole_PROTAGONIST {
			// 对主角隐藏真实角色，显示为未指定。
			playerViewChar.Role = model.RoleType_ROLE_UNKNOWN
		} else {
			playerViewChar.Role = charCopy.HiddenRole
		}
		view.Characters[id] = playerViewChar
	}

	// 过滤玩家信息
	view.Players = make(map[int32]*model.PlayerViewPlayer)
	for id, p := range ge.GameState.Players {
		playerCopy := proto.Clone(p).(*model.Player)
		playerViewPlayer := &model.PlayerViewPlayer{
			Id:   id,
			Name: playerCopy.Name,
			Role: playerCopy.Role,
		}
		view.Players[id] = playerViewPlayer
	}

	// 添加玩家特定信息
	view.YourHand = player.Hand
	if player.Role == model.PlayerRole_PROTAGONIST {
		view.YourDeductions = player.DeductionKnowledge
	}

	return view
}
