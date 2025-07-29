package phase

import (
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// CardEffectsPhase 卡牌效果阶段
type CardEffectsPhase struct{ basePhase }

// Type 返回阶段类型
func (p *CardEffectsPhase) Type() model.GamePhase { return model.GamePhase_CARD_RESOLVE }

// Enter 进入阶段
func (p *CardEffectsPhase) Enter(ge GameEngine) Phase {
	resolver := NewMovementResolver()
	charMovements := resolver.CalculateMovements(ge.GetGameState().PlayedCardsThisDay)

	// 应用计算出的移动
	for charID, movement := range charMovements {
		if movement.Forbidden {
			continue
		}

		char := ge.GetCharacterByID(charID)
		if char == nil || !char.IsAlive {
			continue
		}

		finalH := movement.H
		finalV := movement.V

		// 对角线移动算作一次水平移动和一次垂直移动
		if movement.D > 0 {
			finalH += movement.D
			finalV += movement.D
		}

		// 水平和垂直移动的组合成为对角线移动
		if finalH > 0 && finalV > 0 {
			finalH--
			finalV--
			// 实际上，我们正在进行一次对角线移动，然后是任何剩余的水平/垂直移动
			ge.MoveCharacter(char, 1, 1) // 对角线
		}

		if finalH > 0 {
			ge.MoveCharacter(char, finalH, 0) // 水平
		}
		if finalV > 0 {
			ge.MoveCharacter(char, 0, finalV) // 垂直
		}
	}

	resolveOtherCards(ge)

	// 卡牌效果结算后，我们可能会进入能力阶段
	return &AbilitiesPhase{}
}

// CharacterMovement 保存角色的计算移动向量
type CharacterMovement struct {
	H         int
	V         int
	D         int
	Forbidden bool
}

// MovementResolver 根据打出的牌计算角色移动
type MovementResolver struct{}

// NewMovementResolver 创建一个新的 MovementResolver
func NewMovementResolver() *MovementResolver {
	return &MovementResolver{}
}

// CalculateMovements 汇总每个角色从打出的牌中获得的移动效果
func (mr *MovementResolver) CalculateMovements(playedCards map[int32]*model.Card) map[int32]CharacterMovement {
	charMovements := make(map[int32]CharacterMovement)

	for _, card := range playedCards {
		targetCharID, isCharTarget := card.Target.(*model.Card_TargetCharacterId)
		if !isCharTarget {
			continue
		}

		movement := charMovements[targetCharID.TargetCharacterId]
		if movement.Forbidden {
			continue // 移动已被禁止，无需进一步计算
		}

		switch card.Config.Type {
		case model.CardType_MOVE_HORIZONTALLY:
			movement.H++
		case model.CardType_MOVE_VERTICALLY:
			movement.V++
		case model.CardType_MOVE_DIAGONALLY:
			movement.D++
		case model.CardType_FORBID_MOVEMENT:
			movement = CharacterMovement{Forbidden: true} // 取消所有移动
		}
		charMovements[targetCharID.TargetCharacterId] = movement
	}
	return charMovements
}

// resolveMovement 处理回合中打出的所有移动牌
func resolveMovement(ge GameEngine) {
	resolver := NewMovementResolver()
	charMovements := resolver.CalculateMovements(ge.GetGameState().PlayedCardsThisDay)

	// 应用计算出的移动
	for charID, movement := range charMovements {
		if movement.Forbidden {
			continue
		}

		char := ge.GetCharacterByID(charID)
		if char == nil || !char.IsAlive {
			continue
		}

		finalH := movement.H
		finalV := movement.V

		// 对角线移动算作一次水平移动和一次垂直移动
		if movement.D > 0 {
			finalH += movement.D
			finalV += movement.D
		}

		// 水平和垂直移动的组合成为对角线移动
		if finalH > 0 && finalV > 0 {
			finalH--
			finalV--
			// 实际上，我们正在进行一次对角线移动，然后是任何剩余的水平/垂直移动
			ge.MoveCharacter(char, 1, 1) // 对角线
			}

			if finalH > 0 {
				ge.MoveCharacter(char, finalH, 0) // 水平
			}
			if finalV > 0 {
				ge.MoveCharacter(char, 0, finalV) // 垂直
		}
	}
}


// resolveOtherCards 处理非移动牌
func resolveOtherCards(ge GameEngine) {
	for _, card := range ge.GetGameState().PlayedCardsThisDay {
		switch card.Config.Type {
		case model.CardType_MOVE_HORIZONTALLY, model.CardType_MOVE_VERTICALLY, model.CardType_MOVE_DIAGONALLY, model.CardType_FORBID_MOVEMENT:
			continue // 已经处理过
		default:
			// TODO: 为其他卡牌类型（妄想、好感等）实现逻辑
			ge.Logger().Info("resolving other card", zap.String("card", card.Config.Name))
		}
	}
}