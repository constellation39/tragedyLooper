package phasehandler

import (
	"time"
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// MastermindCardPlayPhase is the phasehandler where the mastermind plays their cards.
type MastermindCardPlayPhase struct {
	basePhase
	mastermindCardsPlayed int
}

// Type returns the phasehandler type.
func (p *MastermindCardPlayPhase) Type() model.GamePhase {
	return model.GamePhase_GAME_PHASE_MASTERMIND_CARD_PLAY
}

// Enter is called when the phasehandler begins.
func (p *MastermindCardPlayPhase) Enter(ge GameEngine) Phase {
	p.mastermindCardsPlayed = 0
	
	// TODO: Trigger Mastermind AI action here.
	return nil
}

// HandleAction handles an action from a player.
func (p *MastermindCardPlayPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if player.Role != model.PlayerRole_PLAYER_ROLE_MASTERMIND {
		return nil
	}

	if payload, ok := action.Payload.(*model.PlayerActionPayload_PlayCard); ok {
		handlePlayCardAction(ge, player, payload.PlayCard)
		p.mastermindCardsPlayed++
	}

	if p.mastermindCardsPlayed >= 3 {
		return GetPhase(model.GamePhase_GAME_PHASE_PROTAGONIST_CARD_PLAY)
	}

	return nil
}

// HandleTimeout handles a timeout.
func (p *MastermindCardPlayPhase) HandleTimeout(ge GameEngine) Phase {
	return GetPhase(model.GamePhase_GAME_PHASE_PROTAGONIST_CARD_PLAY)
}

// TimeoutDuration returns the timeout duration for this phasehandler.
func (p *MastermindCardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second }

func init() {
	RegisterPhase(&MastermindCardPlayPhase{})
}
