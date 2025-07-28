package engine

import (
	model "tragedylooper/internal/game/proto/v1"

	"go.uber.org/zap"
)

// --- Player Action Handlers ---

func (ge *GameEngine) handlePlayerAction(playerID int32, action *model.PlayerActionPayload) {
	player, ok := ge.GameState.Players[playerID]
	if !ok {
		ge.logger.Warn("Action from unknown player", zap.Int32("playerID", playerID))
		return
	}

	ge.logger.Info("Handling player action", zap.String("player", player.Name), zap.Any("action", action.Payload))

	switch p := action.Payload.(type) {
	case *model.PlayerActionPayload_PlayCard:
		ge.handlePlayCardAction(player, p.PlayCard)
	case *model.PlayerActionPayload_UseAbility:
		ge.handleUseAbilityAction(player, p.UseAbility)
	case *model.PlayerActionPayload_MakeGuess:
		ge.handleMakeGuessAction(p.MakeGuess)
	case *model.PlayerActionPayload_ChooseOption:
		// TODO: Handle ChooseOption
	default:
		ge.logger.Warn("Unknown action type", zap.Any("action", action.Payload))
	}
}

func (ge *GameEngine) handlePlayCardAction(player *model.Player, payload *model.PlayCardPayload) {

	var playedCard *model.Card
	cardFound := false
	for i, card := range player.Hand {
		if card.Config.Id == payload.CardId {
			if card.Config.OncePerLoop && card.UsedThisLoop {
				ge.logger.Warn("Attempted to play a card that was already used this loop", zap.Int32("cardID", card.Config.Id))
				return // Card already used
			}
			playedCard = card
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...) // Remove card from hand
			cardFound = true
			break
		}
	}

	if !cardFound {
		ge.logger.Warn("Attempted to play a card not in hand", zap.Int32("cardID", payload.CardId), zap.Int32("playerID", player.Id))
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

	if _, ok := ge.GameState.PlayedCardsThisDay[player.Id]; !ok {
		ge.GameState.PlayedCardsThisDay[player.Id] = playedCard
	}
	if _, ok := ge.GameState.PlayedCardsThisLoop[playedCard.Config.Id]; !ok {
		ge.GameState.PlayedCardsThisLoop[playedCard.Config.Id] = true
	}

	ge.playerReady[player.Id] = true
}

func (ge *GameEngine) handleUseAbilityAction(_ *model.Player, payload *model.UseAbilityPayload) {

	var ability *model.Ability
	abilityFound := false
	char, ok := ge.GameState.Characters[payload.CharacterId]
	if !ok {
		ge.logger.Warn("Character not found for ability use", zap.Int32("characterID", payload.CharacterId))
		return
	}

	for i := range char.Abilities {
		if char.Abilities[i].Config.Id == payload.AbilityId {
			ability = char.Abilities[i]
			abilityFound = true
			break
		}
	}

	if !abilityFound {
		ge.logger.Warn("Ability not found on character", zap.Int32("abilityID", payload.AbilityId), zap.Int32("characterID", payload.CharacterId))
		return
	}

	if err := ge.applyEffect(ability.Config.Effect, ability, payload, nil); err != nil {
		ge.logger.Error("Failed to apply effect for ability", zap.String("abilityName", ability.Config.Name), zap.Error(err))
		return
	}

	if ability.Config.OncePerLoop {
		ability.UsedThisLoop = true
	}
}

func (ge *GameEngine) handleMakeGuessAction(payload *model.MakeGuessPayload) {

	correctGuesses := 0
	totalCharactersToGuess := 0
	for charID, guessedRole := range payload.GuessedRoles {
		char, exists := ge.GameState.Characters[charID]
		if !exists {
			continue // Ignore guesses for non-existent characters
		}
		// Only count characters that have a hidden role to be guessed
		if char.HiddenRole != model.RoleType_ROLE_TYPE_UNSPECIFIED {
			totalCharactersToGuess++
			if char.HiddenRole == guessedRole {
				correctGuesses++
			}
		}
	}

	if totalCharactersToGuess > 0 && correctGuesses == totalCharactersToGuess {
		ge.endGame(model.PlayerRole_PROTAGONIST)
	} else {
		ge.endGame(model.PlayerRole_MASTERMIND)
	}
}
