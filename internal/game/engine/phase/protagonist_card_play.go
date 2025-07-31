package phase

import (
	"time"
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// ProtagonistCardPlayPhase is the phase where the protagonists play their cards.
type ProtagonistCardPlayPhase struct {
	basePhase
	currentPlayerIndex int
}

// Type returns the phase type.
func (p *ProtagonistCardPlayPhase) Type() model.GamePhase {
	return model.GamePhase_GAME_PHASE_PROTAGONIST_CARD_PLAY
}

// Enter is called when the phase begins.
func (p *ProtagonistCardPlayPhase) Enter(ge GameEngine) Phase {
	p.currentPlayerIndex = 0
	ge.ResetPlayerReadiness()
	// TODO: Trigger Protagonist AI action here.
	return nil
}

// HandleAction handles an action from a player.
func (p *ProtagonistCardPlayPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	protagonists := ge.GetProtagonistPlayers()
	if len(protagonists) == 0 {
		return GetPhase(model.GamePhase_GAME_PHASE_CARD_REVEAL)
	}

	if player.Role != model.PlayerRole_PLAYER_ROLE_PROTAGONIST || player.Id != protagonists[p.currentPlayerIndex].Id {
		ge.Logger().Warn("Received action from player out of turn", zap.String("expected_player", protagonists[p.currentPlayerIndex].Name), zap.String("actual_player", player.Name))
		return nil
	}

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_PlayCard:
		handlePlayCardAction(ge, player, payload.PlayCard)
	case *model.PlayerActionPayload_PassTurn:
		handlePassTurnAction(ge, player)
	}

	ge.SetPlayerReady(player.Id)
	p.currentPlayerIndex++

	if p.currentPlayerIndex >= len(protagonists) {
		return GetPhase(model.GamePhase_GAME_PHASE_CARD_REVEAL)
	}

	// TODO: Trigger AI for the next protagonist.
	return nil
}

// HandleTimeout handles a timeout.
func (p *ProtagonistCardPlayPhase) HandleTimeout(ge GameEngine) Phase {
	return GetPhase(model.GamePhase_GAME_PHASE_CARD_REVEAL)
}

// TimeoutDuration returns the timeout duration for this phase.
func (p *ProtagonistCardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second }

func init() {
	RegisterPhase(&ProtagonistCardPlayPhase{})
}
