package phasehandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// ProtagonistCardPlayPhase is the phase where the protagonists play their cards.
type ProtagonistCardPlayPhase struct {
	BasePhase
	protagonistTurnIndex int
}

// Type returns the phase type.
func (p *ProtagonistCardPlayPhase) Type() model.GamePhase {
	return model.GamePhase_GAME_PHASE_PROTAGONIST_CARD_PLAY
}

// Enter is called when the phase begins.
func (p *ProtagonistCardPlayPhase) Enter(ge GameEngine) {
	p.protagonistTurnIndex = 0
	protagonists := ge.GetProtagonistPlayers()
	if len(protagonists) > 0 {
		ge.RequestAIAction(protagonists[0].Id)
	}
}

// HandleAction handles an action from a player.
func (p *ProtagonistCardPlayPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) bool {
	protagonists := ge.GetProtagonistPlayers()
	if len(protagonists) == 0 {
		return true
	}

	if player.Role != model.PlayerRole_PLAYER_ROLE_PROTAGONIST || player.Id != protagonists[p.protagonistTurnIndex].Id {
		ge.Logger().Warn("Received action from player out of turn", zap.String("expected_player", protagonists[p.protagonistTurnIndex].Name), zap.String("actual_player", player.Name))
		return false
	}

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_PlayCard:
		handlePlayCardAction(ge, player, payload.PlayCard)
	case *model.PlayerActionPayload_PassTurn:
		handlePassTurnAction(ge, player)
	}

	p.protagonistTurnIndex++

	if p.protagonistTurnIndex >= len(protagonists) {
		return true
	}

	// Trigger AI for the next protagonist.
	nextProtagonist := protagonists[p.protagonistTurnIndex]
	ge.RequestAIAction(nextProtagonist.Id)
	return false
}

func init() {
	RegisterPhase(&ProtagonistCardPlayPhase{})
}
