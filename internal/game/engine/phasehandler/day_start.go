package phasehandler

import (
	"time"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// DayStartPhase 天开始阶段
type DayStartPhase struct{}

// HandleAction is the default implementation for Phase interface, does nothing and returns nil.
func (p *DayStartPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	return nil
}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *DayStartPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }

// HandleTimeout is the default implementation for Phase interface, does nothing and returns nil.
func (p *DayStartPhase) HandleTimeout(ge GameEngine) Phase { return nil }

// Exit is the default implementation for Phase interface, does nothing.
func (p *DayStartPhase) Exit(ge GameEngine) {}

// TimeoutDuration is the default implementation for Phase interface, returns 0, indicating no timeout.
func (p *DayStartPhase) TimeoutDuration() time.Duration { return 0 }

func (p *DayStartPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_DAY_START }
func (p *DayStartPhase) Enter(ge GameEngine) {
	ge.GetGameState().CurrentDay++
	ge.GetGameState().PlayedCardsThisDay = make(map[int32]*model.CardList)

	ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_DAY_ADVANCED, &model.EventPayload{
		Payload: &model.EventPayload_DayAdvanced{DayAdvanced: &model.DayAdvancedEvent{Day: ge.GetGameState().CurrentDay, Loop: ge.GetGameState().CurrentLoop}},
	})
}

func init() {
	RegisterPhase(&DayStartPhase{})
}
