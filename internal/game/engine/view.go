package engine

import (
	"google.golang.org/protobuf/proto"

	model "tragedylooper/pkg/proto/v1"
)

// generatePlayerView creates a filtered view of the game state for a specific player.
// This method is NOT thread-safe and must only be called from within the runGameLoop goroutine.
func (ge *GameEngine) GeneratePlayerView(playerID int32) *model.PlayerView {
	player := ge.GameState.Players[playerID]
	if player == nil {
		return &model.PlayerView{} // Or handle error
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

	// Filter characters based on player role
	view.Characters = make(map[int32]*model.PlayerViewCharacter)
	for id, char := range ge.GameState.Characters {
		// Create a copy to avoid races and unintended modification of the core state.
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
			// Hide the true role from Protagonists, showing it as unspecified.
			// The Mastermind will see the true roles.
			// charCopy.HiddenRole = model.RoleType_ROLE_TYPE_UNSPECIFIED // This field does not exist on PlayerViewCharacter
		}
		view.Characters[id] = playerViewChar
	}

	// Filter player info
	view.Players = make(map[int32]*model.PlayerViewPlayer)
	for id, p := range ge.GameState.Players {
		playerCopy := proto.Clone(p).(*model.Player)
		playerViewPlayer := &model.PlayerViewPlayer{
			Id:   id,
			Name: playerCopy.Name,
			Role: playerCopy.Role,
		}
		// playerCopy.Hand = nil // Hide other players' hands - not applicable to PlayerViewPlayer
		view.Players[id] = playerViewPlayer
	}

	// Add player-specific info
	view.YourHand = player.Hand
	if player.Role == model.PlayerRole_PROTAGONIST {
		view.YourDeductions = player.DeductionKnowledge
	}

	return view
}
