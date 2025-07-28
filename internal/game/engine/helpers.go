package engine

import model "tragedylooper/internal/game/proto/v1"

func (ge *GameEngine) getMastermindPlayer() *model.Player {
	for _, p := range ge.GameState.Players {
		if p.Role == model.PlayerRole_MASTERMIND {
			return p
		}
	}
	return nil
}

func (ge *GameEngine) isMastermindReady() bool {
	mastermind := ge.getMastermindPlayer()
	if mastermind == nil {
		ge.logger.Error("No mastermind player found")
		return false // Or handle as a critical error
	}
	return ge.playerReady[mastermind.Id]
}

func (ge *GameEngine) areAllPlayersReady() bool {
	for playerID := range ge.GameState.Players {
		if !ge.playerReady[playerID] {
			return false
		}
	}
	return true
}