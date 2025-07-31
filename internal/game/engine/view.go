package engine

import (
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

	// Filter character information based on player role
	view.Characters = make(map[int32]*model.PlayerViewCharacter, len(ge.GameState.Characters))
	for id, char := range ge.GameState.Characters {
		playerViewChar := &model.PlayerViewCharacter{
			Id:              id,
			Name:            char.Config.Name,
			Traits:          char.Traits,
			CurrentLocation: char.CurrentLocation,
			Paranoia:        char.Paranoia,
			Goodwill:        char.Goodwill,
			Intrigue:        char.Intrigue,
			Abilities:       char.Abilities,
			IsAlive:         char.IsAlive,
			InPanicMode:     char.InPanicMode,
			Rules:           char.Config.Rules,
		}
		if player.Role == model.PlayerRole_PLAYER_ROLE_PROTAGONIST {
			// Hide the true role from protagonists, show as unknown.
			playerViewChar.Role = model.RoleType_ROLE_TYPE_ROLE_UNKNOWN
		} else {
			playerViewChar.Role = char.HiddenRole
		}
		view.Characters[id] = playerViewChar
	}

	// Filter player information
	view.Players = make(map[int32]*model.PlayerViewPlayer, len(ge.GameState.Players))
	for id, p := range ge.GameState.Players {
		view.Players[id] = &model.PlayerViewPlayer{
			Id:   id,
			Name: p.Name,
			Role: p.Role,
		}
	}

	// Add player-specific information
	view.YourHand = player.Hand.Cards
	if player.Role == model.PlayerRole_PLAYER_ROLE_PROTAGONIST {
		view.YourDeductions = player.DeductionKnowledge
	}

	return view
}
