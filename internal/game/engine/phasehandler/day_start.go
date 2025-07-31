package phasehandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// DayStartPhase 天开始阶段
type DayStartPhase struct{ basePhase }

func (p *DayStartPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_DAY_START }
func (p *DayStartPhase) Enter(ge GameEngine) Phase {
	ge.GetGameState().CurrentDay++
	ge.GetGameState().PlayedCardsThisDay = make(map[int32]*model.CardList)
	
	ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_DAY_ADVANCED, &model.EventPayload{
		Payload: &model.EventPayload_DayAdvanced{DayAdvanced: &model.DayAdvancedEvent{Day: ge.GetGameState().CurrentDay, Loop: ge.GetGameState().CurrentLoop}},
	})
	return GetPhase(model.GamePhase_GAME_PHASE_MASTERMIND_CARD_PLAY)
}

func init() {
	RegisterPhase(&DayStartPhase{})
}
