package phase

import (
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// DayEndPhase is where end-of-day checks are performed.
type DayEndPhase struct{ basePhase }

// Type returns the phase type.
func (p *DayEndPhase) Type() model.GamePhase { return model.GamePhase_DAY_END }

// Enter is called when the phase begins.
func (p *DayEndPhase) Enter(ge GameEngine) Phase {
	logger := ge.Logger().Named("DayEndPhase")
	script := ge.GetGameRepo().GetScript()

	// 1. Check for loop loss conditions
	for _, endCond := range script.LoseConditions {
		if endCond.Type == model.EndConditionType_PROTAGONIST_GUESS_FAIL {
			for _, req := range endCond.Requirements {
				met, err := ge.CheckCondition(req)
				if err != nil {
					logger.Error("Error checking loop loss condition", zap.Error(err))
					continue
				}
				if met {
					logger.Info("Loop loss condition met", zap.String("description", endCond.Description))
					ge.ApplyAndPublishEvent(model.GameEventType_LOOP_LOSS, &model.EventPayload{})
					return &LoopEndPhase{}
				}
			}
		}
	}

	// 2. Check for protagonist win conditions (e.g., all loss conditions prevented)
	// This logic can be complex. A simple version checks if all incidents that are part of loss conditions were prevented.
	// For now, we will assume a simple check.

	// 3. If no win/loss, check if it's the last day of the loop
	if ge.GetGameState().CurrentDay >= ge.GetGameState().DaysPerLoop {
		logger.Info("End of loop reached by day count")
		return &LoopEndPhase{}
	}

	// 4. Otherwise, proceed to the next day
	logger.Info("Proceeding to next day")
	return &DayStartPhase{}
}
