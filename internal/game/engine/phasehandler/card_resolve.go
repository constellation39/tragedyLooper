package phasehandler

import (
	"sort"

	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// CardResolvePhase is the phase where the effects of played cards are resolved.
type CardResolvePhase struct {
	BasePhase
}

// Type 返回阶段类型。
func (p *CardResolvePhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_CARD_RESOLVE }

// Enter 在阶段开始时调用。
func (p *CardResolvePhase) Enter(ge GameEngine) PhaseState {
	logger := ge.Logger().Named("CardResolvePhase")
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
	return PhaseComplete
}

func (p *CardResolvePhase) resolveForbidMovement(logger *zap.Logger, cards []*model.Card) map[int32]bool {
	forbidden := make(map[int32]bool)
	for _, card := range cards {
		if card.Config.Type == model.CardType_CARD_TYPE_FORBID_MOVEMENT {
			if target, ok := card.Target.(*model.Card_TargetCharacterId); ok {
				logger.Info("Character movement forbidden", zap.Int32("charID", target.TargetCharacterId), zap.String("card", card.Config.Name))
				forbidden[target.TargetCharacterId] = true
			}
		}
	}
	return forbidden
}

func (p *CardResolvePhase) resolveMovement(logger *zap.Logger, ge GameEngine, cards []*model.Card, forbiddenMoves map[int32]bool) {
	movements := make(map[int32]struct{ H, V, D int })

	for _, card := range cards {
		if target, ok := card.Target.(*model.Card_TargetCharacterId); ok {
			charID := target.TargetCharacterId
			if forbiddenMoves[charID] {
				continue
			}

			move := movements[charID]
			switch card.Config.Type {
			case model.CardType_CARD_TYPE_MOVE_HORIZONTALLY:
				move.H++
			case model.CardType_CARD_TYPE_MOVE_VERTICALLY:
				move.V++
			case model.CardType_CARD_TYPE_MOVE_DIAGONALLY:
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

func (p *CardResolvePhase) resolveStatEffects(logger *zap.Logger, ge GameEngine, cards []*model.Card) {
	forbidParanoiaInc, forbidGoodwillInc, forbidIntrigueInc := p.gatherForbidEffects(cards)

	for _, card := range cards {
		if target, ok := card.Target.(*model.Card_TargetCharacterId); ok {
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

func (p *CardResolvePhase) gatherForbidEffects(cards []*model.Card) (map[int32]bool, map[int32]bool, map[int32]bool) {
	forbidParanoiaInc := make(map[int32]bool)
	forbidGoodwillInc := make(map[int32]bool)
	forbidIntrigueInc := make(map[int32]bool)

	for _, card := range cards {
		if target, ok := card.Target.(*model.Card_TargetCharacterId); ok {
			charID := target.TargetCharacterId
			switch card.Config.Type {
			case model.CardType_CARD_TYPE_FORBID_PARANOIA_INCREASE:
				forbidParanoiaInc[charID] = true
			case model.CardType_CARD_TYPE_FORBID_GOODWILL_INCREASE:
				forbidGoodwillInc[charID] = true
			case model.CardType_CARD_TYPE_FORBID_INTRIGUE_INCREASE:
				forbidIntrigueInc[charID] = true
			}
		}
	}
	return forbidParanoiaInc, forbidGoodwillInc, forbidIntrigueInc
}

func (p *CardResolvePhase) applyStatEffect(logger *zap.Logger, ge GameEngine, charID int32, cardType model.CardType, amount int32, forbidParanoia, forbidGoodwill, forbidIntrigue map[int32]bool) {
	switch cardType {
	case model.CardType_CARD_TYPE_ADD_PARANOIA:
		if forbidParanoia[charID] && amount > 0 {
			logger.Info("Paranoia increase forbidden", zap.Int32("charID", charID))
			return
		}
		ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_PARANOIA_ADJUSTED, &model.EventPayload{
			Payload: &model.EventPayload_ParanoiaAdjusted{ParanoiaAdjusted: &model.ParanoiaAdjustedEvent{
				CharacterId: charID,
				Amount:      amount,
			}},
		})
	case model.CardType_CARD_TYPE_ADD_GOODWILL:
		if forbidGoodwill[charID] && amount > 0 {
			logger.Info("Goodwill increase forbidden", zap.Int32("charID", charID))
			return
		}
		ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_GOODWILL_ADJUSTED, &model.EventPayload{
			Payload: &model.EventPayload_GoodwillAdjusted{GoodwillAdjusted: &model.GoodwillAdjustedEvent{
				CharacterId: charID,
				Amount:      amount,
			}},
		})
	case model.CardType_CARD_TYPE_ADD_INTRIGUE:
		if forbidIntrigue[charID] && amount > 0 {
			logger.Info("Intrigue increase forbidden", zap.Int32("charID", charID))
			return
		}
		ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_INTRIGUE_ADJUSTED, &model.EventPayload{
			Payload: &model.EventPayload_IntrigueAdjusted{IntrigueAdjusted: &model.IntrigueAdjustedEvent{
				CharacterId: charID,
				Amount:      amount,
			}},
		})
	}
}

// getAllPlayedCards 将已打出卡牌的映射扁平化为单个切片并对其进行排序。
// 排序对于确保确定性的解析顺序很重要。
func getAllPlayedCards(ge GameEngine) []*model.Card {
	var cards []*model.Card
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

func init() {
	RegisterPhase(&CardResolvePhase{})
}
