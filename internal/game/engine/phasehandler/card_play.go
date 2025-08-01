package phasehandler

import (
	"time"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

type CardPlayTurn int

const (
	MastermindCardTurn CardPlayTurn = iota
	ProtagonistCardTurn
)

// CardPlayPhase is the phase where players play their cards.
type CardPlayPhase struct {
	turn                  CardPlayTurn
	mastermindCardsPlayed int
	protagonistTurnIndex  int
}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *CardPlayPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }

// Exit is the default implementation for Phase interface, does nothing.
func (p *CardPlayPhase) Exit(ge GameEngine) {}

// NewCardPlayPhase creates a new CardPlayPhase.
func NewCardPlayPhase(turn CardPlayTurn) Phase {
	return &CardPlayPhase{turn: turn}
}

// Type returns the phase type.
func (p *CardPlayPhase) Type() model.GamePhase {
	if p.turn == MastermindCardTurn {
		return model.GamePhase_GAME_PHASE_MASTERMIND_CARD_PLAY
	}
	return model.GamePhase_GAME_PHASE_PROTAGONIST_CARD_PLAY
}

// Enter is called when the phase begins.
func (p *CardPlayPhase) Enter(ge GameEngine) Phase {
	if p.turn == MastermindCardTurn {
		p.mastermindCardsPlayed = 0
		ge.RequestAIAction(ge.GetMastermindPlayer().Id)
	} else {
		p.protagonistTurnIndex = 0
		protagonists := ge.GetProtagonistPlayers()
		if len(protagonists) > 0 {
			ge.RequestAIAction(protagonists[0].Id)
		}
	}
	return nil
}

// HandleAction handles an action from a player.
func (p *CardPlayPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if p.turn == MastermindCardTurn {
		return p.handleMastermindAction(ge, player, action)
	}
	return p.handleProtagonistAction(ge, player, action)
}

// HandleTimeout handles a timeout.
func (p *CardPlayPhase) HandleTimeout(ge GameEngine) Phase {
	if p.turn == MastermindCardTurn {
		return GetPhase(model.GamePhase_GAME_PHASE_PROTAGONIST_CARD_PLAY)
	}
	return GetPhase(model.GamePhase_GAME_PHASE_CARD_REVEAL)
}

// TimeoutDuration returns the timeout duration for this phase.
func (p *CardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second }

func (p *CardPlayPhase) handleMastermindAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
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

func (p *CardPlayPhase) handleProtagonistAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	protagonists := ge.GetProtagonistPlayers()
	if len(protagonists) == 0 {
		return GetPhase(model.GamePhase_GAME_PHASE_CARD_REVEAL)
	}

	if player.Role != model.PlayerRole_PLAYER_ROLE_PROTAGONIST || player.Id != protagonists[p.protagonistTurnIndex].Id {
		ge.Logger().Warn("Received action from player out of turn", zap.String("expected_player", protagonists[p.protagonistTurnIndex].Name), zap.String("actual_player", player.Name))
		return nil
	}

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_PlayCard:
		handlePlayCardAction(ge, player, payload.PlayCard)
	case *model.PlayerActionPayload_PassTurn:
		handlePassTurnAction(ge, player)
	}

	p.protagonistTurnIndex++

	if p.protagonistTurnIndex >= len(protagonists) {
		return GetPhase(model.GamePhase_GAME_PHASE_CARD_REVEAL)
	}

	// Trigger AI for the next protagonist.
	nextProtagonist := protagonists[p.protagonistTurnIndex]
	ge.RequestAIAction(nextProtagonist.Id)

	return nil
}

func init() {
	RegisterPhase(NewCardPlayPhase(MastermindCardTurn))
	RegisterPhase(NewCardPlayPhase(ProtagonistCardTurn))
}
