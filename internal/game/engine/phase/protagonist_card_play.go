package phase

import (
	"time"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// ProtagonistCardPlayPhase is the phase where the Protagonists play their cards.
type ProtagonistCardPlayPhase struct {
	basePhase
	currentPlayerIndex int
}

// Type returns the phase type.
func (p *ProtagonistCardPlayPhase) Type() model.GamePhase {
	return model.GamePhase_CARD_PLAY
}

// Enter is called when the phase begins.
func (p *ProtagonistCardPlayPhase) Enter(ge GameEngine) Phase {
	p.currentPlayerIndex = 0
	ge.ResetPlayerReadiness()
	// Potentially trigger AI action for the first protagonist here.
	return nil
}

// HandleAction handles an action from a player.
func (p *ProtagonistCardPlayPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	protagonists := ge.GetProtagonistPlayers()
	if len(protagonists) == 0 {
		return &CardRevealPhase{}
	}

	if player.Id != protagonists[p.currentPlayerIndex].Id {
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
		return &CardRevealPhase{}
	}

	// Potentially trigger AI for the next protagonist.
	return nil
}

// HandleTimeout handles a timeout.
func (p *ProtagonistCardPlayPhase) HandleTimeout(ge GameEngine) Phase {
	// Handle timeout, maybe play a random card or pass the turn.
	protagonists := ge.GetProtagonistPlayers()
	if len(protagonists) == 0 {
		return &CardRevealPhase{}
	}

	// Pass the turn for the current player
	ge.SetPlayerReady(protagonists[p.currentPlayerIndex].Id)
	p.currentPlayerIndex++

	if p.currentPlayerIndex >= len(protagonists) {
		return &CardRevealPhase{}
	}

	// Potentially trigger AI for the next protagonist.
	return nil
}

// TimeoutDuration returns the timeout duration for this phase.
func (p *ProtagonistCardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second }

// handlePlayCardAction handles a protagonist playing a card.
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

	dayState, ok := ge.GetGameState().PlayedCardsThisDay[player.Id]
	if !ok {
		dayState = &model.CardList{}
		ge.GetGameState().PlayedCardsThisDay[player.Id] = dayState
	}
	dayState.Cards = append(dayState.Cards, playedCard)

	// Mark the card as used for this loop
	ge.GetGameState().PlayedCardsThisLoop[playedCard.Config.Id] = true

	// Apply the card's effects
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

// handlePassTurnAction handles a player passing their turn.
func handlePassTurnAction(ge GameEngine, player *model.Player) {
	ge.Logger().Info("Player passed turn", zap.String("player", player.Name))
}