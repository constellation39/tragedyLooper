package phasehandler

import (
	"time"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// CardResolvePhase 卡牌结算阶段，在此阶段处理已打出卡牌的效果。
type CardResolvePhase struct{}

// HandleAction is the default implementation for Phase interface, does nothing and returns nil.
func (p *CardResolvePhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) {
}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *CardResolvePhase) HandleEvent(ge GameEngine, event *model.GameEvent) {}

// HandleTimeout is the default implementation for Phase interface, does nothing and returns nil.
func (p *CardResolvePhase) HandleTimeout(ge GameEngine) {}

// Exit is the default implementation for Phase interface, does nothing.
func (p *CardResolvePhase) Exit(ge GameEngine) {}

// TimeoutDuration is the default implementation for Phase interface, returns 0, indicating no timeout.
func (p *CardResolvePhase) TimeoutDuration() time.Duration { return 0 }

// Type 返回阶段类型，表示当前是卡牌结算阶段。
func (p *CardResolvePhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_CARD_RESOLVE }

// Enter 进入卡牌结算阶段。
func (p *CardResolvePhase) Enter(ge GameEngine) {}

func init() {
	RegisterPhase(&CardResolvePhase{})
}
