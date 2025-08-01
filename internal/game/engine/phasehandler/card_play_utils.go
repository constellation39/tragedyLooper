package phasehandler

import (
	"fmt"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// handlePlayCardAction handles the common logic for playing a card.
func handlePlayCardAction(ge GameEngine, player *model.Player, payload *model.PlayCardPayload) {
	playedCard, err := takeCardFromPlayer(player, payload.CardId)
	if err != nil {
		ge.Logger().Warn("Failed to play card", zap.Error(err), zap.Int32("cardID", payload.CardId), zap.Int32("playerID", player.Id))
		return
	}

	// Add target information to the card instance before storing it
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

	// Apply card effects
	if playedCard.Config.Effect != nil {
		abilityPayload := &model.UseAbilityPayload{}
		switch t := payload.Target.(type) {
		case *model.PlayCardPayload_TargetCharacterId:
			abilityPayload.Target = &model.UseAbilityPayload_TargetCharacterId{TargetCharacterId: t.TargetCharacterId}
		case *model.PlayCardPayload_TargetLocation:
			abilityPayload.Target = &model.UseAbilityPayload_TargetLocation{TargetLocation: t.TargetLocation}
		}

		for _, effect := range playedCard.Config.Effect.SubEffects {
			err := ge.ApplyEffect(effect, nil, abilityPayload, nil)
			if err != nil {
				ge.Logger().Error("Failed to apply card effect", zap.Error(err))
			}
		}
	}
}

// handlePassTurnAction handles the action of a player passing their turn.
func handlePassTurnAction(ge GameEngine, player *model.Player) {
	ge.Logger().Info("Player passed turn", zap.String("player", player.Name))
}

// takeCardFromPlayer finds a card in the player's hand, removes it, and returns it.
func takeCardFromPlayer(player *model.Player, cardID int32) (*model.Card, error) {
	for i, card := range player.Hand.Cards {
		if card.Config.Id == cardID {
			// Remove the card from the hand and return it
			player.Hand.Cards = append(player.Hand.Cards[:i], player.Hand.Cards[i+1:]...)
			return card, nil
		}
	}
	return nil, fmt.Errorf("card %d not found in player's hand", cardID)
}
