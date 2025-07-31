package phasehandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// CardResolvePhase 卡牌结算阶段，在此阶段处理已打出卡牌的效果。
type CardResolvePhase struct{ basePhase }

// Type 返回阶段类型，表示当前是卡牌结算阶段。
func (p *CardResolvePhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_CARD_EFFECTS }

// Enter 进入卡牌结算阶段。
func (p *CardResolvePhase) Enter(ge GameEngine) Phase {
	return GetPhase(model.GamePhase_GAME_PHASE_INCIDENTS)
}

func init() {
	RegisterPhase(&CardResolvePhase{})
}
