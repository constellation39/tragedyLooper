package phase

import (
	"time"
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// ProtagonistCardPlayPhase 是主角出牌的阶段。
type ProtagonistCardPlayPhase struct {
	basePhase
	currentPlayerIndex int
}

// Type 返回阶段类型。
func (p *ProtagonistCardPlayPhase) Type() model.GamePhase {
	return model.GamePhase_GAME_PHASE_CARD_PLAY
}

// Enter 在阶段开始时调用。
func (p *ProtagonistCardPlayPhase) Enter(ge GameEngine) Phase {
	p.currentPlayerIndex = 0
	ge.ResetPlayerReadiness()
	// 可以在此处为第一个主角触发 AI 行动。
	return nil
}

// HandleAction 处理来自玩家的行动。
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

	// 可能会为下一个主角触发 AI。
	return nil
}

// HandleTimeout 处理超时。
func (p *ProtagonistCardPlayPhase) HandleTimeout(ge GameEngine) Phase {
	// 处理超时，可以随机出牌或跳过回合。
	protagonists := ge.GetProtagonistPlayers()
	if len(protagonists) == 0 {
		return &CardRevealPhase{}
	}

	// 为当前玩家跳过回合
	ge.SetPlayerReady(protagonists[p.currentPlayerIndex].Id)
	p.currentPlayerIndex++

	if p.currentPlayerIndex >= len(protagonists) {
		return &CardRevealPhase{}
	}

	// 可能会为下一个主角触发 AI。
	return nil
}

// TimeoutDuration 返回此阶段的超时持续时间。
func (p *ProtagonistCardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second }

// handlePlayCardAction 处理主角出牌的动作。
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
