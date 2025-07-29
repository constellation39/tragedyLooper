package engine

import (
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/v1"

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
		ge.handleMakeGuessAction(player, p.MakeGuess)
	case *model.PlayerActionPayload_PassTurn:
		ge.handlePassTurnAction(player)
	case *model.PlayerActionPayload_ChooseOption:
		// TODO: Handle ChooseOption
	default:
		ge.logger.Warn("Unknown action type", zap.Any("action", action.Payload))
	}
}

func (ge *GameEngine) handlePlayCardAction(player *model.Player, payload *model.PlayCardPayload) {
	if ge.currentPhase.Type() != model.GamePhase_CARD_PLAY {
		ge.logger.Warn("player tried to play card outside of the card play phase", zap.Int32("player_id", player.Id), zap.String("phase", ge.currentPhase.Type().String()))
		return
	}

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

	if _, ok := ge.GameState.PlayedCardsThisDay[player.Id]; ok {
		ge.logger.Warn("player tried to play a second card in one day", zap.Int32("player_id", player.Id))
		// Potentially return the card to the hand or handle it as a misplay.
	}
	ge.GameState.PlayedCardsThisDay[player.Id] = playedCard

	// Mark card as used for the loop
	ge.GameState.PlayedCardsThisLoop[playedCard.Config.Id] = true

	ge.playerReady[player.Id] = true
}

func (ge *GameEngine) handleUseAbilityAction(player *model.Player, payload *model.UseAbilityPayload) {
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
	// Note: Using an ability does not automatically make a player "ready".
	// They must explicitly pass their turn with PassTurnAction.
}

func (ge *GameEngine) handleMakeGuessAction(player *model.Player, payload *model.MakeGuessPayload) {
	if ge.currentPhase.Type() != model.GamePhase_PROTAGONIST_GUESS {
		ge.logger.Warn("player tried to make a guess outside of the guess phase", zap.Int32("player_id", player.Id), zap.String("phase", ge.currentPhase.Type().String()))
		return
	}

	// For now, we assume the first protagonist to guess ends the game.
	if player.Role != model.PlayerRole_PROTAGONIST {
		ge.logger.Warn("non-protagonist player tried to make a guess", zap.Int32("player_id", player.Id))
		return
	}

	script := ge.gameConfig.GetScript()
	if script == nil {
		ge.logger.Error("failed to get script to verify guess")
		ge.endGame(model.PlayerRole_MASTERMIND) // End game, mastermind wins by default on error
		return
	}

	correctGuesses := 0
	for _, roleInfo := range script.Characters {
		if guessedRole, ok := payload.GuessedRoles[roleInfo.CharacterId]; ok {
			if guessedRole == roleInfo.HiddenRole {
				correctGuesses++
			}
		}
	}

	if correctGuesses == len(script.Characters) {
		ge.endGame(model.PlayerRole_PROTAGONIST)
	} else {
		ge.endGame(model.PlayerRole_MASTERMIND)
	}
}

func (ge *GameEngine) handlePassTurnAction(player *model.Player) {
	ge.logger.Info("Player passed turn", zap.String("player", player.Name))
	ge.playerReady[player.Id] = true
}

func (ge *GameEngine) dealInitialCards() {
	script := ge.gameConfig.GetScript()
	if script == nil {
		ge.logger.Error("cannot deal cards, script not loaded")
		return
	}

	mastermind := ge.getMastermindPlayer()
	if mastermind != nil {
		for _, cardID := range script.MastermindCardIds {
			cardConfig, err := loader.Get[*model.CardConfig](ge.gameConfig, cardID)
			if err != nil {
				ge.logger.Warn("mastermind card config not found", zap.Int32("cardID", cardID))
				continue
			}
			mastermind.Hand = append(mastermind.Hand, &model.Card{Config: cardConfig})
		}
	}

	protagonists := ge.getProtagonistPlayers()
	for _, protagonist := range protagonists {
		for _, cardID := range script.ProtagonistCardIds {
			cardConfig, err := loader.Get[*model.CardConfig](ge.gameConfig, cardID)
			if err != nil {
				ge.logger.Warn("protagonist card config not found", zap.Int32("cardID", cardID))
				continue
			}
			protagonist.Hand = append(protagonist.Hand, &model.Card{Config: cardConfig})
		}
	}
}
