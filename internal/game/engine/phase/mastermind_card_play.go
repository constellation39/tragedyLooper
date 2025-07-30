package phase

import (
	"fmt"
	"time"
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// MastermindCardPlayPhase 是主谋出牌的阶段。
type MastermindCardPlayPhase struct {
	basePhase
	cardsPlayed int
}

// Type 返回阶段类型。
func (p *MastermindCardPlayPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_CARD_PLAY }

// Enter 在阶段开始时调用。
func (p *MastermindCardPlayPhase) Enter(ge GameEngine) Phase {
	p.cardsPlayed = 0
	// 可以在此处触发主谋的 AI 行动。
	return nil
}

// HandleAction 处理来自玩家的行动。
func (p *MastermindCardPlayPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if player.Role != model.PlayerRole_PLAYER_ROLE_MASTERMIND {
		ge.Logger().Warn("Received action from non-mastermind player during MastermindCardPlayPhase", zap.String("player", player.Name))
		return nil
	}

	if payload, ok := action.Payload.(*model.PlayerActionPayload_PlayCard); ok {
		p.handlePlayCardAction(ge, player, payload.PlayCard)
	}

	if p.cardsPlayed >= 3 {
		return &ProtagonistCardPlayPhase{}
	}

	return nil
}

// HandleTimeout 处理超时。
func (p *MastermindCardPlayPhase) HandleTimeout(ge GameEngine) Phase {
	// 处理超时，可以为主谋随机出牌。
	return &ProtagonistCardPlayPhase{}
}

// TimeoutDuration 返回此阶段的超时持续时间。
func (p *MastermindCardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second }

func (p *MastermindCardPlayPhase) handlePlayCardAction(ge GameEngine, player *model.Player, payload *model.PlayCardPayload) {
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

	p.cardsPlayed++
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
