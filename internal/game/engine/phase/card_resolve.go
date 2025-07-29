package phase

import (
	model "tragedylooper/pkg/proto/v1"
)

// CardResolvePhase 卡牌结算阶段
type CardResolvePhase struct{ basePhase }

// Type 返回阶段类型
func (p *CardResolvePhase) Type() model.GamePhase { return model.GamePhase_CARD_RESOLVE }

// Enter 进入阶段
func (p *CardResolvePhase) Enter(ge GameEngine) Phase {
	return &CardEffectsPhase{}
}