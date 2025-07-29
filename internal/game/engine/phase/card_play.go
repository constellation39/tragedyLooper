package phase // 定义游戏阶段包

import (
	"fmt"
	"time"
	model "tragedylooper/pkg/proto/v1" // 导入协议缓冲区模型

	"go.uber.org/zap" // 导入 Zap 日志库
)

// CardPlayPhase 卡牌打出阶段，玩家在此阶段打出卡牌。
type CardPlayPhase struct{ basePhase }

// Type 返回阶段类型，表示当前是卡牌打出阶段。
func (p *CardPlayPhase) Type() model.GamePhase { return model.GamePhase_CARD_PLAY }

// Enter 进入卡牌打出阶段。
// ge: 游戏引擎接口。
// 返回值: 下一个阶段的实例（如果立即切换）。
func (p *CardPlayPhase) Enter(ge GameEngine) Phase {
	// 玩家有一定的时间打出他们的牌。
	return nil
}

// HandleAction 处理玩家在卡牌打出阶段的操作。
// ge: 游戏引擎接口。
// playerID: 执行操作的玩家ID。
// action: 玩家操作的负载。
// 返回值: 如果阶段发生变化，则返回新的阶段实例；否则返回 nil。
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

	// 如果所有玩家都已准备好，则切换到卡牌揭示阶段。
	if ge.AreAllPlayersReady() {
		return &CardRevealPhase{}
	}

	return nil
}

// HandleTimeout 处理卡牌打出阶段的超时。
// ge: 游戏引擎接口。
// 返回值: 下一个阶段的实例。
func (p *CardPlayPhase) HandleTimeout(ge GameEngine) Phase {
	// 如果玩家没有及时行动，我们可能会为他们自动跳过回合。
	return &CardRevealPhase{}
}

// HandleEvent 处理卡牌打出阶段接收到的游戏事件。
// ge: 游戏引擎接口。
// event: 接收到的游戏事件。
// 返回值: 如果阶段发生变化，则返回新的阶段实例；否则返回 nil。
func (p *CardPlayPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase {
	if ge.AreAllPlayersReady() {
		return &CardRevealPhase{}
	}
	return nil
}

// TimeoutDuration 返回卡牌打出阶段的超时持续时间。
func (p *CardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second } // 示例超时

// handlePlayCardAction 处理玩家打出卡牌的操作。
// ge: 游戏引擎接口。
// player: 执行操作的玩家。
// payload: 打出卡牌的负载信息。
func handlePlayCardAction(ge GameEngine, player *model.Player, payload *model.PlayCardPayload) {
	playedCard, err := takeCardFromPlayer(player, payload.CardId)
	if err != nil {
		ge.Logger().Warn("Failed to play card", zap.Error(err), zap.Int32("cardID", payload.CardId), zap.Int32("playerID", player.Id))
		return
	}

	// 在存储之前将目标信息添加到卡牌实例中
	switch t := payload.Target.(type) {
	case *model.PlayCardPayload_TargetCharacterId:
		playedCard.Target = &model.Card_TargetCharacterId{TargetCharacterId: t.TargetCharacterId}
	case *model.PlayCardPayload_TargetLocation:
		playedCard.Target = &model.Card_TargetLocation{TargetLocation: t.TargetLocation}
	}
	playedCard.UsedThisLoop = true // 标记为已使用

	if _, ok := ge.GetGameState().PlayedCardsThisDay[player.Id]; ok {
		ge.Logger().Warn("player tried to play a second card in one day", zap.Int32("player_id", player.Id))
		// 可能会将卡牌退回到手牌或将其作为误操作处理。
	}
	ge.GetGameState().PlayedCardsThisDay[player.Id] = playedCard

	// 将卡牌标记为本循环已使用
	ge.GetGameState().PlayedCardsThisLoop[playedCard.Config.Id] = true

	ge.SetPlayerReady(player.Id)
}

// takeCardFromPlayer 在玩家手牌中找到一张牌，将其移除并返回。
// player: 玩家对象。
// cardID: 要移除的卡牌ID。
// 返回值: 移除的卡牌对象和可能发生的错误。
func takeCardFromPlayer(player *model.Player, cardID int32) (*model.Card, error) {
	for i, card := range player.Hand {
		if card.Config.Id == cardID {
			if card.Config.OncePerLoop && card.UsedThisLoop {
				return nil, fmt.Errorf("card %d was already used this loop", cardID)
			}
			// 从手牌中移除卡牌并返回
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			return card, nil
		}
	}
	return nil, fmt.Errorf("card %d not found in player's hand", cardID)
}

// handlePassTurnAction 处理玩家跳过回合的操作。
// ge: 游戏引擎接口。
// player: 执行操作的玩家。
func handlePassTurnAction(ge GameEngine, player *model.Player) {
	ge.Logger().Info("Player passed turn", zap.String("player", player.Name))
	ge.SetPlayerReady(player.Id)
}
