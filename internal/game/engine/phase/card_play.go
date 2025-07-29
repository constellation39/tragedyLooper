package phase

import (
	"fmt"
	"time"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// --- CardPlayPhase ---
type CardPlayPhase struct{ basePhase }

func (p *CardPlayPhase) Type() model.GamePhase { return model.GamePhase_CARD_PLAY }
func (p *CardPlayPhase) Enter(ge GameEngine) Phase {
	// Players have a certain amount of time to play their cards.
	return nil
}
func (p *CardPlayPhase) HandleAction(ge GameEngine, playerID int32, action *model.PlayerActionPayload) Phase {
	state := ge.GetGameState()
	player, ok := state.Players[playerID]
	if !ok {
		ge.Logger().Warn("Action from unknown player", zap.Int32("playerID", playerID))
		return nil
	}

	ge.Logger().Info("Handling player action", zap.String("player", player.Name), zap.Any("action", action.Payload))

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_PlayCard:
		handlePlayCardAction(ge, player, payload.PlayCard)
	case *model.PlayerActionPayload_PassTurn:
		handlePassTurnAction(ge, player)
	}

	if ge.AreAllPlayersReady() {
		return &CardRevealPhase{}
	}

	return nil
}

func (p *CardPlayPhase) HandleTimeout(ge GameEngine) Phase {
	// If players don't act in time, we might auto-pass for them.
	return &CardRevealPhase{}
}
func (p *CardPlayPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase {
	if ge.AreAllPlayersReady() {
		return &CardRevealPhase{}
	}
	return nil
}
func (p *CardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second } // Example timeout

func handlePlayCardAction(ge GameEngine, player *model.Player, payload *model.PlayCardPayload) {
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

	if _, ok := ge.GetGameState().PlayedCardsThisDay[player.Id]; ok {
		ge.Logger().Warn("player tried to play a second card in one day", zap.Int32("player_id", player.Id))
		// Potentially return the card to the hand or handle it as a misplay.
	}
	ge.GetGameState().PlayedCardsThisDay[player.Id] = playedCard

	// Mark card as used for the loop
	ge.GetGameState().PlayedCardsThisLoop[playedCard.Config.Id] = true

	ge.SetPlayerReady(player.Id)
}

// takeCardFromPlayer finds a card in a player's hand, removes it, and returns it.
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

func handlePassTurnAction(ge GameEngine, player *model.Player) {
	ge.Logger().Info("Player passed turn", zap.String("player", player.Name))
	ge.SetPlayerReady(player.Id)
}
