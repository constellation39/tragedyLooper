package engine

import (
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"tragedylooper/internal/game/proto/model"
)

// --- Player Action Handlers ---

func (ge *GameEngine) handlePlayerAction(action *model.PlayerAction) {
	player, ok := ge.GameState.Players[action.PlayerId]
	if !ok {
		ge.logger.Warn("Action from unknown player", zap.Int32("playerID", action.PlayerId))
		return
	}

	ge.logger.Info("Handling player action", zap.String("player", player.Name), zap.String("actionType", string(action.Type)))

	switch action.Type {
	case model.ActionType_ACTION_TYPE_PLAY_CARD:
		ge.handlePlayCardAction(player, action)
	case model.ActionType_ACTION_TYPE_USE_ABILITY:
		ge.handleUseAbilityAction(player, action)
	case model.ActionType_ACTION_TYPE_MAKE_GUESS:
		ge.handleMakeGuessAction(action)
	case model.ActionType_ACTION_TYPE_READY_FOR_NEXT_PHASE:
		ge.handleReadyForNextPhaseAction(player)
	default:
		ge.logger.Warn("Unknown action type", zap.String("actionType", string(action.Type)))
	}
}

func (ge *GameEngine) handlePlayCardAction(player *model.Player, action *model.PlayerAction) {
	var payload model.PlayCardPayload
	if err := protojson.Unmarshal(action.Payload.MarshalJSON(), &payload); err != nil {
		ge.logger.Error("Failed to unmarshal PlayCardPayload", zap.Error(err))
		return
	}

	var playedCard *model.Card
	cardFound := false
	for i, card := range player.Hand {
		if card.Id == payload.CardId {
			if card.OncePerLoop && card.UsedThisLoop {
				ge.logger.Warn("Attempted to play a card that was already used this loop", zap.Int32("cardID", card.Id))
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
	playedCard.Target = payload.Target
	playedCard.UsedThisLoop = true // Mark as used

	if _, ok := ge.GameState.PlayedCardsThisDay[player.Id]; !ok {
		ge.GameState.PlayedCardsThisDay[player.Id] = &model.CardList{}
	}
	if _, ok := ge.GameState.PlayedCardsThisLoop[player.Id]; !ok {
		ge.GameState.PlayedCardsThisLoop[player.Id] = &model.CardList{}
	}

	ge.GameState.PlayedCardsThisDay[player.Id].Cards = append(ge.GameState.PlayedCardsThisDay[player.Id].Cards, playedCard)
	ge.GameState.PlayedCardsThisLoop[player.Id].Cards = append(ge.GameState.PlayedCardsThisLoop[player.Id].Cards, playedCard)
	ge.playerReady[player.Id] = true
}

func (ge *GameEngine) handleUseAbilityAction(player *model.Player, action *model.PlayerAction) {
	var payload model.UseAbilityPayload
	if err := action.Payload.UnmarshalTo(&payload); err != nil {
		ge.logger.Error("Failed to unmarshal UseAbilityPayload", zap.Error(err))
		return
	}

	var ability *model.Ability
	abilityFound := false
	char, ok := ge.GameState.Characters[payload.CharacterId]
	if !ok {
		ge.logger.Warn("Character not found for ability use", zap.Int32("characterID", payload.CharacterId))
		return
	}

	for i := range char.Abilities {
		if char.Abilities[i].Id == payload.AbilityId {
			ability = char.Abilities[i]
			abilityFound = true
			break
		}
	}

	if !abilityFound {
		ge.logger.Warn("Ability not found on character", zap.Int32("abilityID", payload.AbilityId), zap.Int32("characterID", payload.CharacterId))
		return
	}

	if err := ge.applyEffect(ability.Effect, ability, &payload); err != nil {
		ge.logger.Error("Failed to apply effect for ability", zap.String("abilityName", ability.Name), zap.Error(err))
		return
	}

	if ability.OncePerLoop {
		ability.UsedThisLoop = true
	}
}

func (ge *GameEngine) handleReadyForNextPhaseAction(player *model.Player) {
	ge.playerReady[player.Id] = true
}

func (ge *GameEngine) handleMakeGuessAction(action *model.PlayerAction) {
	if ge.GameState.CurrentPhase != model.GamePhase_GAME_PHASE_PROTAGONIST_GUESS {
		ge.logger.Warn("MakeGuess action received outside of the guess phase")
		return
	}

	var payload model.MakeGuessPayload
	if err := protojson.Unmarshal(action.Payload.MarshalJSON(), &payload); err != nil {
		ge.logger.Error("Failed to unmarshal MakeGuessPayload", zap.Error(err))
		return
	}

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
		ge.endGame(model.PlayerRole_PLAYER_ROLE_PROTAGONIST)
	} else {
		ge.endGame(model.PlayerRole_PLAYER_ROLE_MASTERMIND)
	}
}
