package phase

import (
	"sort"
	v1 "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// CardEffectsPhase 是解析已打出卡牌效果的阶段。
type CardEffectsPhase struct{ basePhase }

// Type 返回阶段类型。
func (p *CardEffectsPhase) Type() v1.GamePhase { return v1.GamePhase_CARD_EFFECTS }

// Enter 在阶段开始时调用。
func (p *CardEffectsPhase) Enter(ge GameEngine) Phase {
	logger := ge.Logger().Named("CardEffectsPhase")
	playedCards := getAllPlayedCards(ge)

	// --- 解析顺序 --- //
	// 1. 禁止移动
	// 2. 移动
	// 3. 其他禁止效果
	// 4. 其他效果（偏执、好感、阴谋）

	logger.Info("Resolving card effects")

	// 步骤 1：解析禁止移动
	forbiddenMoves := p.resolveForbidMovement(logger, playedCards)

	// 步骤 2：解析移动
	p.resolveMovement(logger, ge, playedCards, forbiddenMoves)

	// 步骤 3 & 4：解析其他效果（包括其禁止效果）
	p.resolveStatEffects(logger, ge, playedCards)

	logger.Info("Finished resolving card effects")

	// 卡牌效果解析后，我们可能会进入能力阶段
	return &AbilitiesPhase{}
}

func (p *CardEffectsPhase) resolveForbidMovement(logger *zap.Logger, cards []*v1.Card) map[int32]bool {
	forbidden := make(map[int32]bool)
	for _, card := range cards {
		if card.Config.Type == v1.CardType_FORBID_MOVEMENT {
			if target, ok := card.Target.(*v1.Card_TargetCharacterId); ok {
				logger.Info("Character movement forbidden", zap.Int32("charID", target.TargetCharacterId), zap.String("card", card.Config.Name))
				forbidden[target.TargetCharacterId] = true
			}
		}
	}
	return forbidden
}

func (p *CardEffectsPhase) resolveMovement(logger *zap.Logger, ge GameEngine, cards []*v1.Card, forbiddenMoves map[int32]bool) {
	movements := make(map[int32]struct{ H, V, D int })

	for _, card := range cards {
		if target, ok := card.Target.(*v1.Card_TargetCharacterId); ok {
			charID := target.TargetCharacterId
			if forbiddenMoves[charID] {
				continue
			}

			move := movements[charID]
			switch card.Config.Type {
			case v1.CardType_MOVE_HORIZONTALLY:
				move.H++
			case v1.CardType_MOVE_VERTICALLY:
				move.V++
			case v1.CardType_MOVE_DIAGONALLY:
				move.D++
			}
			movements[charID] = move
		}
	}

	for charID, move := range movements {
		char := ge.GetCharacterByID(charID)
		if char == nil || !char.IsAlive {
			continue
		}

		// 简化的移动逻辑：应用对角线，然后是水平，然后是垂直
		// 此逻辑可能需要根据组合移动的具体游戏规则进行调整。
		if move.D > 0 {
			ge.MoveCharacter(char, move.D, move.D)
		}
		if move.H > 0 {
			ge.MoveCharacter(char, move.H, 0)
		}
		if move.V > 0 {
			ge.MoveCharacter(char, 0, move.V)
		}
		logger.Info("Character moved", zap.Int32("charID", charID), zap.Any("movement", move))
	}
}

func (p *CardEffectsPhase) resolveStatEffects(logger *zap.Logger, ge GameEngine, cards []*v1.Card) {
	forbidParanoiaInc, forbidGoodwillInc, forbidIntrigueInc := p.gatherForbidEffects(cards)

	for _, card := range cards {
		if target, ok := card.Target.(*v1.Card_TargetCharacterId); ok {
			charID := target.TargetCharacterId
			char := ge.GetCharacterByID(charID)
			if char == nil {
				continue
			}

			var amount int32 = 1 // 默认数量
			p.applyStatEffect(logger, ge, charID, card.Config.Type, amount, forbidParanoiaInc, forbidGoodwillInc, forbidIntrigueInc)
		}
	}
}

func (p *CardEffectsPhase) gatherForbidEffects(cards []*v1.Card) (map[int32]bool, map[int32]bool, map[int32]bool) {
	forbidParanoiaInc := make(map[int32]bool)
	forbidGoodwillInc := make(map[int32]bool)
	forbidIntrigueInc := make(map[int32]bool)

	for _, card := range cards {
		if target, ok := card.Target.(*v1.Card_TargetCharacterId); ok {
			charID := target.TargetCharacterId
			switch card.Config.Type {
			case v1.CardType_CARD_TYPE_FORBID_PARANOIA_INCREASE:
				forbidParanoiaInc[charID] = true
			case v1.CardType_CARD_TYPE_FORBID_GOODWILL_INCREASE:
				forbidGoodwillInc[charID] = true
			case v1.CardType_CARD_TYPE_FORBID_INTRIGUE_INCREASE:
				forbidIntrigueInc[charID] = true
			}
		}
	}
	return forbidParanoiaInc, forbidGoodwillInc, forbidIntrigueInc
}

func (p *CardEffectsPhase) applyStatEffect(logger *zap.Logger, ge GameEngine, charID int32, cardType v1.CardType, amount int32, forbidParanoia, forbidGoodwill, forbidIntrigue map[int32]bool) {
	switch cardType {
	case v1.CardType_CARD_TYPE_ADD_PARANOIA:
		if forbidParanoia[charID] && amount > 0 {
			logger.Info("Paranoia increase forbidden", zap.Int32("charID", charID))
			return
		}
		ge.ApplyAndPublishEvent(v1.GameEventType_GAME_EVENT_TYPE_PARANOIA_ADJUSTED, &v1.EventPayload{
			Payload: &v1.EventPayload_ParanoiaAdjusted{ParanoiaAdjusted: &v1.ParanoiaAdjustedEvent{
				CharacterId: charID,
				Amount:      amount,
			}},
		})
	case v1.CardType_CARD_TYPE_ADD_GOODWILL:
		if forbidGoodwill[charID] && amount > 0 {
			logger.Info("Goodwill increase forbidden", zap.Int32("charID", charID))
			return
		}
		ge.ApplyAndPublishEvent(v1.GameEventType_GAME_EVENT_TYPE_GOODWILL_ADJUSTED, &v1.EventPayload{
			Payload: &v1.EventPayload_GoodwillAdjusted{GoodwillAdjusted: &v1.GoodwillAdjustedEvent{
				CharacterId: charID,
				Amount:      amount,
			}},
		})
	case v1.CardType_CARD_TYPE_ADD_INTRIGUE:
		if forbidIntrigue[charID] && amount > 0 {
			logger.Info("Intrigue increase forbidden", zap.Int32("charID", charID))
			return
		}
		ge.ApplyAndPublishEvent(v1.GameEventType_GAME_EVENT_TYPE_INTRIGUE_ADJUSTED, &v1.EventPayload{
			Payload: &v1.EventPayload_IntrigueAdjusted{IntrigueAdjusted: &v1.IntrigueAdjustedEvent{
				CharacterId: charID,
				Amount:      amount,
			}},
		})
	}
}

// getAllPlayedCards 将已打出卡牌的映射扁平化为单个切片并对其进行排序。
// 排序对于确保确定性的解析顺序很重要。
func getAllPlayedCards(ge GameEngine) []*v1.Card {
	var cards []*v1.Card
	for _, cardList := range ge.GetGameState().PlayedCardsThisDay {
		cards = append(cards, cardList.Cards...)
	}

	// 按确定性键（例如，卡牌 ID）对卡牌进行排序。
	// 这确保了游戏实例之间的解析顺序是一致的。
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Config.Id < cards[j].Config.Id
	})

	return cards
}
