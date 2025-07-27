package engine

import (
	"google.golang.org/protobuf/proto"

	model "tragedylooper/internal/game/proto/v1"
)

// generatePlayerView creates a filtered view of the game state for a specific player.
// This method is NOT thread-safe and must only be called from within the runGameLoop goroutine.
func (ge *GameEngine) generatePlayerView(playerID int32) *model.PlayerView {
	player := ge.GameState.Players[playerID]
	if player == nil {
		return &model.PlayerView{} // Or handle error
	}

	view := &model.PlayerView{
		GameId:             ge.GameState.GameId,
		ScriptId:           ge.GameState.Script.Id,
		CurrentDay:         ge.GameState.CurrentDay,
		CurrentLoop:        ge.GameState.CurrentLoop,
		CurrentPhase:       ge.GameState.CurrentPhase,
		ActiveTragedies:    ge.GameState.ActiveTragedies,
		PreventedTragedies: ge.GameState.PreventedTragedies,
		PublicEvents:       ge.GameState.DayEvents,
	}

	// Filter characters based on player role
	view.Characters = make(map[int32]*model.Character)
	for id, char := range ge.GameState.Characters {
		// Create a copy to avoid races and unintended modification of the core state.
		charCopy := proto.Clone(char).(*model.Character)
		if player.Role == model.PlayerRole_PLAYER_ROLE_PROTAGONIST {
			// Hide the true role from Protagonists, showing it as unspecified.
			// The Mastermind will see the true roles.
			charCopy.HiddenRole = model.RoleType_ROLE_TYPE_UNSPECIFIED
		}
		view.Characters[id] = charCopy
	}

	// Filter player info
	view.Players = make(map[int32]*model.Player)
	for id, p := range ge.GameState.Players {
		playerCopy := proto.Clone(p).(*model.Player)
		if id != playerID {
			playerCopy.Hand = nil // Hide other players' hands
		}
		view.Players[id] = playerCopy
	}

	// Add player-specific info
	view.YourHand = player.Hand
	if player.Role == model.PlayerRole_PLAYER_ROLE_PROTAGONIST {
		view.YourDeductions = player.DeductionKnowledge
	}

	return view
}
