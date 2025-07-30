package phase

import (
	"sort"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// CardEffectsPhase is where the effects of played cards are resolved.
type CardEffectsPhase struct{ basePhase }

// Type returns the phase type.
func (p *CardEffectsPhase) Type() model.GamePhase { return model.GamePhase_CARD_EFFECTS }

// Enter is called when the phase begins.
func (p *CardEffectsPhase) Enter(ge GameEngine) Phase {
	logger := ge.Logger().Named("CardEffectsPhase")
	playedCards := getAllPlayedCards(ge)

	// --- Resolution Order --- //
	// 1. Forbid Movement
	// 2. Movement
	// 3. Other Forbid Effects
	// 4. Other Effects (Paranoia, Goodwill, Intrigue)

	logger.Info("Resolving card effects")

	// Step 1: Resolve Forbid Movement
	forbiddenMoves := p.resolveForbidMovement(logger, playedCards)

	// Step 2: Resolve Movement
	p.resolveMovement(logger, ge, playedCards, forbiddenMoves)

	// Step 3 & 4: Resolve other effects (including their forbids)
	p.resolveStatEffects(logger, ge, playedCards)

	logger.Info("Finished resolving card effects")

	// After card effects are resolved, we might go to the ability phase
	return &AbilitiesPhase{}
}

func (p *CardEffectsPhase) resolveForbidMovement(logger *zap.Logger, cards []*model.Card) map[int32]bool {
	forbidden := make(map[int32]bool)
	for _, card := range cards {
		if card.Config.Type == model.CardType_FORBID_MOVEMENT {
			if target, ok := card.Target.(*model.Card_TargetCharacterId); ok {
				logger.Info("Character movement forbidden", zap.Int32("charID", target.TargetCharacterId), zap.String("card", card.Config.Name))
				forbidden[target.TargetCharacterId] = true
			}
		}
	}
	return forbidden
}

func (p *CardEffectsPhase) resolveMovement(logger *zap.Logger, ge GameEngine, cards []*model.Card, forbiddenMoves map[int32]bool) {
	movements := make(map[int32]struct{ H, V, D int })

	for _, card := range cards {
		if target, ok := card.Target.(*model.Card_TargetCharacterId); ok {
			charID := target.TargetCharacterId
			if forbiddenMoves[charID] {
				continue
			}

			move := movements[charID]
			switch card.Config.Type {
			case model.CardType_MOVE_HORIZONTALLY:
				move.H++
			case model.CardType_MOVE_VERTICALLY:
				move.V++
			case model.CardType_MOVE_DIAGONALLY:
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

		// Simplified movement logic: apply diagonal, then horizontal, then vertical
		// This logic might need to be adjusted based on specific game rules for combining movements.
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

func (p *CardEffectsPhase) resolveStatEffects(logger *zap.Logger, ge GameEngine, cards []*model.Card) {
	// Step 3: Gather all forbid effects for stats
	forbidParanoiaInc := make(map[int32]bool)
	forbidGoodwillInc := make(map[int32]bool)
	forbidIntrigueInc := make(map[int32]bool)

	for _, card := range cards {
		if target, ok := card.Target.(*model.Card_TargetCharacterId); ok {
			charID := target.TargetCharacterId
			switch card.Config.Type {
			case model.CardType_FORBID_PARANOIA_INCREASE:
				forbidParanoiaInc[charID] = true
			case model.CardType_FORBID_GOODWILL_INCREASE:
				forbidGoodwillInc[charID] = true
			case model.CardType_FORBID_INTRIGUE_INCREASE:
				forbidIntrigueInc[charID] = true
			}
		}
	}

	// Step 4: Apply stat adjustments
	for _, card := range cards {
		if target, ok := card.Target.(*model.Card_TargetCharacterId); ok {
			charID := target.TargetCharacterId
			char := ge.GetCharacterByID(charID)
			if char == nil {
				continue
			}

			var amount int32 = 1 // Default amount, can be specified on the card later

			switch card.Config.Type {
			case model.CardType_ADD_PARANOIA:
				if forbidParanoiaInc[charID] && amount > 0 {
					logger.Info("Paranoia increase forbidden", zap.Int32("charID", charID))
					continue
				}
				ge.ApplyAndPublishEvent(model.GameEventType_PARANOIA_ADJUSTED, &model.EventPayload{
					Payload: &model.EventPayload_ParanoiaAdjusted{ParanoiaAdjusted: &model.ParanoiaAdjustedEvent{
						CharacterId: charID,
						Amount:      amount,
					}},
				})
			case model.CardType_ADD_GOODWILL:
				if forbidGoodwillInc[charID] && amount > 0 {
					logger.Info("Goodwill increase forbidden", zap.Int32("charID", charID))
					continue
				}
				ge.ApplyAndPublishEvent(model.GameEventType_GOODWILL_ADJUSTED, &model.EventPayload{
					Payload: &model.EventPayload_GoodwillAdjusted{GoodwillAdjusted: &model.GoodwillAdjustedEvent{
						CharacterId: charID,
						Amount:      amount,
					}},
				})
			case model.CardType_ADD_INTRIGUE:
				if forbidIntrigueInc[charID] && amount > 0 {
					logger.Info("Intrigue increase forbidden", zap.Int32("charID", charID))
					continue
				}
				ge.ApplyAndPublishEvent(model.GameEventType_INTRIGUE_ADJUSTED, &model.EventPayload{
					Payload: &model.EventPayload_IntrigueAdjusted{IntrigueAdjusted: &model.IntrigueAdjustedEvent{
						CharacterId: charID,
						Amount:      amount,
					}},
				})
			}
		}
	}
}

// getAllPlayedCards flattens the map of played cards into a single slice and sorts them.
// The sorting is important to ensure a deterministic resolution order.
func getAllPlayedCards(ge GameEngine) []*model.Card {
	var cards []*model.Card
	for _, cardList := range ge.GetGameState().PlayedCardsThisDay {
		cards = append(cards, cardList.Cards...)
	}

	// Sort cards by a deterministic key, e.g., Card ID.
	// This ensures that the resolution order is consistent between game instances.
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Config.Id < cards[j].Config.Id
	})

	return cards
}
