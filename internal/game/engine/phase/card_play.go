package phase

import (
	"fmt"
	"time"
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// CardPlayPhase 是主谋和主角出牌的统一阶段。
type CardPlayPhase struct {
	basePhase
	mastermindCardsPlayed int
	protagonistPlayerIndex int
}

// Type 返回阶段类型。
func (p *CardPlayPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_CARD_PLAY }

// Enter 在阶段开始时调用。
func (p *CardPlayPhase) Enter(ge GameEngine) Phase {
	p.mastermindCardsPlayed = 0
	p.protagonistPlayerIndex = 0
	ge.ResetPlayerReadiness()
	// 可以在此处触发主谋的 AI 行动。
	return nil
}

// HandleAction 处理来自玩家的行动。
func (p *CardPlayPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if player.Role == model.PlayerRole_PLAYER_ROLE_MASTERMIND {
		return p.handleMastermindAction(ge, player, action)
	} else if player.Role == model.PlayerRole_PLAYER_ROLE_PROTAGONIST {
		return p.handleProtagonistAction(ge, player, action)
	}
	return nil
}

// HandleTimeout 处理超时。
func (p *CardPlayPhase) HandleTimeout(ge GameEngine) Phase {
    // It's complex to determine whose turn it is to timeout.
    // For now, we transition to the next phase.
    return GetPhase(model.GamePhase_GAME_PHASE_CARD_REVEAL)
}


// TimeoutDuration 返回此阶段的超时持续时间。
func (p *CardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second }

func init() {
	RegisterPhase(&CardPlayPhase{})
}

func (p *CardPlayPhase) handleMastermindAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if payload, ok := action.Payload.(*model.PlayerActionPayload_PlayCard); ok {
		handlePlayCardAction(ge, player, payload.PlayCard)
		p.mastermindCardsPlayed++
	}

	if p.mastermindCardsPlayed >= 3 {
		// 主谋出牌结束，轮到主角
		return GetPhase(model.GamePhase_GAME_PHASE_CARD_PLAY) // Re-enter to handle protagonists
	}

	return nil
}

func (p *CardPlayPhase) handleProtagonistAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	protagonists := ge.GetProtagonistPlayers()
	if len(protagonists) == 0 {
		return GetPhase(model.GamePhase_GAME_PHASE_CARD_REVEAL)
	}

	if player.Id != protagonists[p.protagonistPlayerIndex].Id {
		ge.Logger().Warn("Received action from player out of turn", zap.String("expected_player", protagonists[p.protagonistPlayerIndex].Name), zap.String("actual_player", player.Name))
		return nil
	}

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_PlayCard:
		handlePlayCardAction(ge, player, payload.PlayCard)
	case *model.PlayerActionPayload_PassTurn:
		handlePassTurnAction(ge, player)
	}

	ge.SetPlayerReady(player.Id)
	p.protagonistPlayerIndex++

	if p.protagonistPlayerIndex >= len(protagonists) {
		return GetPhase(model.GamePhase_GAME_PHASE_CARD_REVEAL)
	}

	// 可能会为下一个主角触发 AI。
	return nil
}

// handlePlayCardAction 处理出牌的通用逻辑。
func handlePlayCardAction(ge GameEngine, player *model.Player, payload *model.PlayCardPayload) {
	playedCard, err := takeCardFromPlayer(player, payload.CardId)
	if err != nil {
		ge.Logger().Warn("Failed to play card", zap.Error(err), zap.Int32("cardID", payload.CardId), zap.Int32("playerID", player.Id))
		return
	}

	// 在存储卡牌实例之前向其添加目标信息
	switch t := payload.Target.(type) {
	case *model.PlayCardPayload_TargetCharacterId:
		playedCard.Target = &model.Card_TargetCharacterId{TargetCharacterId: t.TargetCharacterId}
	case *model.PlayCardPayload_TargetLocation:
		playedCard.Target = &model.Card_TargetLocation{TargetLocation: t.TargetLocation}
	}
	playedCard.UsedThisLoop = true // 标记为已使用

	dayState, ok := ge.GetGameState().PlayedCardsThisDay[player.Id]
	if !ok {
		dayState = &model.CardList{}
		ge.GetGameState().PlayedCardsThisDay[player.Id] = dayState
	}
	dayState.Cards = append(dayState.Cards, playedCard)

	// 将卡牌标记为本循环已使用
	ge.GetGameState().PlayedCardsThisLoop[playedCard.Config.Id] = true

	// 应用卡牌效果
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

// handlePassTurnAction 处理玩家跳过回合的动作。
func handlePassTurnAction(ge GameEngine, player *model.Player) {
	ge.Logger().Info("Player passed turn", zap.String("player", player.Name))
}

// takeCardFromPlayer 从玩家手牌中找到一张牌，将其移除并返回。
func takeCardFromPlayer(player *model.Player, cardID int32) (*model.Card, error) {
	for i, card := range player.Hand.Cards {
		if card.Config.Id == cardID {
			// 从手牌中移除卡牌并返回
			player.Hand.Cards = append(player.Hand.Cards[:i], player.Hand.Cards[i+1:]...)
			return card, nil
		}
	}
	return nil, fmt.Errorf("card %d not found in player's hand", cardID)
}
