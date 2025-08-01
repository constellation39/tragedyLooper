package phasehandler

import (
	"time"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// MastermindCardPlayPhase is the phase where the mastermind plays their cards.
type MastermindCardPlayPhase struct {
	BasePhase
	mastermindCardsPlayed int
}

// Type returns the phase type.
func (p *MastermindCardPlayPhase) Type() model.GamePhase {
	return model.GamePhase_GAME_PHASE_MASTERMIND_CARD_PLAY
}

// Enter is called when the phase begins.
func (p *MastermindCardPlayPhase) Enter(ge GameEngine) {
	p.mastermindCardsPlayed = 0
	ge.RequestAIAction(ge.GetMastermindPlayer().Id)
}

// HandleAction handles an action from a player.
func (p *MastermindCardPlayPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) bool {
	if player.Role != model.PlayerRole_PLAYER_ROLE_MASTERMIND {
		return false
	}

	if payload, ok := action.Payload.(*model.PlayerActionPayload_PlayCard); ok {
		handlePlayCardAction(ge, player, payload.PlayCard)
		p.mastermindCardsPlayed++
	}
	return p.mastermindCardsPlayed >= 1
}

// TimeoutDuration returns the timeout duration for this phase.
func (p *MastermindCardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second }

func init() {
	RegisterPhase(&MastermindCardPlayPhase{})
}
