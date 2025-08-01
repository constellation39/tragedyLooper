package phasehandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

func handlePlayCardAction(ge GameEngine, player *model.Player, payload *model.PlayCardPayload) {
	gs := ge.GetGameState()

	// Find the card in the player's hand
	var card *model.Card
	cardIndex := -1
	for i, c := range player.Hand.Cards {
		if c.Config.Id == payload.CardId {
			card = c
			cardIndex = i
			break
		}
	}

	if card == nil {
		ge.Logger().Warn("Player tried to play a card they don't have", zap.Int32("player_id", player.Id), zap.Int32("card_id", payload.CardId))
		return
	}

	if _, ok := gs.PlayedCardsThisLoop[card.Config.Id]; ok && card.Config.Type != model.CardType_CARD_TYPE_UNSPECIFIED {
		ge.Logger().Warn("Player tried to play a card that has already been played this loop", zap.Int32("player_id", player.Id), zap.String("card_name", card.Config.Name))
		return
	}

	// Add card to played cards for the day
	if _, ok := gs.PlayedCardsThisDay[player.Id]; !ok {
		gs.PlayedCardsThisDay[player.Id] = &model.CardList{}
	}
	gs.PlayedCardsThisDay[player.Id].Cards = append(gs.PlayedCardsThisDay[player.Id].Cards, card)

	// Mark as played this loop
	gs.PlayedCardsThisLoop[card.Config.Id] = true

	// Remove card from hand
	player.Hand.Cards = append(player.Hand.Cards[:cardIndex], player.Hand.Cards[cardIndex+1:]...)

	ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_CARD_PLAYED, &model.EventPayload{
		Payload: &model.EventPayload_CardPlayed{CardPlayed: &model.CardPlayedEvent{PlayerId: player.Id, Card: card}},
	})
}

func handlePassTurnAction(ge GameEngine, player *model.Player) {
	ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_PLAYER_ACTION, &model.EventPayload{
		Payload: &model.EventPayload_PlayerActionTaken{PlayerActionTaken: &model.PlayerActionTakenEvent{PlayerId: player.Id, Action: &model.PlayerActionPayload{Payload: &model.PlayerActionPayload_PassTurn{PassTurn: &model.PassTurnAction{}}}}},
	})
}
