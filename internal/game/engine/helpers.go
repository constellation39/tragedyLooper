package engine

import model "tragedylooper/internal/game/proto/v1"

func (ge *GameEngine) getProtagonistPlayers() []*model.Player {
	var protagonists []*model.Player
	for _, p := range ge.GameState.Players {
		if p.Role == model.PlayerRole_PROTAGONIST {
			protagonists = append(protagonists, p)
		}
	}
	return protagonists
}

func (ge *GameEngine) getMastermindPlayer() *model.Player {
	for _, p := range ge.GameState.Players {
		if p.Role == model.PlayerRole_MASTERMIND {
			return p
		}
	}
	return nil
}

func (ge *GameEngine) isMastermindReady() bool {
	mm := ge.getMastermindPlayer()
	if mm == nil {
		return true // No mastermind, so they are vacuously ready
	}
	return ge.playerReady[mm.Id]
}

func (ge *GameEngine) areAllPlayersReady() bool {
	for _, p := range ge.GameState.Players {
		if !ge.playerReady[p.Id] {
			return false
		}
	}
	return true
}

func (ge *GameEngine) checkConditions(conditions []*model.Condition, player *model.Player, payload *model.UseAbilityPayload, ability *model.Ability) bool {
	return true
}

func (ge *GameEngine) checkGameEndConditions() (bool, model.PlayerRole) {
	return false, model.PlayerRole_PLAYER_ROLE_UNSPECIFIED
}
