package phasehandler

import (
	
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// CardResolvePhase 卡牌结算阶段，在此阶段处理已打出卡牌的效果。
type CardResolvePhase struct{
	BasePhase
}



// Type 返回阶段类型，表示当前是卡牌结算阶段。
func (p *CardResolvePhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_CARD_RESOLVE }

// Enter 进入卡牌结算阶段。
func (p *CardResolvePhase) Enter(ge GameEngine) {}

func init() {
	RegisterPhase(&CardResolvePhase{})
}
