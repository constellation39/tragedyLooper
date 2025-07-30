package phase

import (
	"fmt"
	"time"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// MastermindCardPlayPhase is the phase where the Mastermind plays their cards.
type MastermindCardPlayPhase struct {
	basePhase
	cardsPlayed int
}

// Type returns the phase type.
func (p *MastermindCardPlayPhase) Type() model.GamePhase { return model.GamePhase_MASTERMIND_CARD_PLAY }

// Enter is called when the phase begins.
func (p *MastermindCardPlayPhase) Enter(ge GameEngine) Phase {
	p.cardsPlayed = 0
	// Potentially trigger AI action for the Mastermind here.
	return nil
}

// HandleAction handles an action from a player.
func (p *MastermindCardPlayPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if player.Role != model.PlayerRole_MASTERMIND {
		ge.Logger().Warn("Received action from non-mastermind player during MastermindCardPlayPhase", zap.String("player", player.Name))
		return nil
	}

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_PlayCard:
		p.handlePlayCardAction(ge, player, payload.PlayCard)
	}

	if p.cardsPlayed >= 3 {
		return &ProtagonistCardPlayPhase{}
	}

	return nil
}

// HandleTimeout handles a timeout.
func (p *MastermindCardPlayPhase) HandleTimeout(ge GameEngine) Phase {
	// Handle timeout, maybe play random cards for the mastermind.
	return &ProtagonistCardPlayPhase{}
}

// TimeoutDuration returns the timeout duration for this phase.
func (p *MastermindCardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second }

func (p *MastermindCardPlayPhase) handlePlayCardAction(ge GameEngine, player *model.Player, payload *model.PlayCardPayload) {
	playedCard, err := takeCardFromPlayer(player, payload.CardId)
	if err != nil {
		ge.Logger().Warn("Failed to play card", zap.Error(err), zap.Int32("cardID", payload.CardId), zap.Int32("playerID", player.Id))
		return
	}

	// Add target info to the card instance before storing it
	switch t := payload.Target.(type) {
	case *model.PlayCardPayload_TargetCharacterId:
		playedCard.Target = &model.Card_TargetCharacterId{TargetCharacterId: t.TargetCharacterId}
	case *model.PlayCardPayload_TargetLocation:
		playedCard.Target = &model.Card_TargetLocation{TargetLocation: t.TargetLocation}
	}
	playedCard.UsedThisLoop = true // Mark as used

	dayState, ok := ge.GetGameState().PlayedCardsThisDay[player.Id]
	if !ok {
		dayState = &model.CardList{}
		ge.GetGameState().PlayedCardsThisDay[player.Id] = dayState
	}
	dayState.Cards = append(dayState.Cards, playedCard)

	// Mark the card as used for this loop
	ge.GetGameState().PlayedCardsThisLoop[playedCard.Config.Id] = true

	p.cardsPlayed++
}

// takeCardFromPlayer finds a card in the player's hand, removes it, and returns it.
func takeCardFromPlayer(player *model.Player, cardID int32) (*model.Card, error) {
	for i, card := range player.Hand {
		if card.Config.Id == cardID {
			if card.Config.OncePerLoop && card.UsedThisLoop {
				return nil, fmt.Errorf("card %d was already used this loop", cardID)
			}
			// Remove card from hand and return it
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			return card, nil
		}
	}
	return nil, fmt.Errorf("card %d not found in player's hand", cardID)
}