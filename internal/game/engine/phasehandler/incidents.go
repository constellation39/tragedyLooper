package phasehandler

import (
	"time"

	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// IncidentsPhase is the phase where the consequences of the day are resolved.
// It advances the day or loop if necessary.

type IncidentsPhase struct{}

func (p *IncidentsPhase) Enter(ge GameEngine) Phase {
	gs := ge.GetGameState()
	script := ge.GetGameRepo().GetScript()

	// Check for day-ending tragedies or other conditions here.
	// For now, we'll just advance the day.

	gs.CurrentDay++
	ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_DAY_ADVANCED, &model.EventPayload{
		Payload: &model.EventPayload_DayAdvanced{DayAdvanced: &model.DayAdvancedEvent{Day: gs.CurrentDay, Loop: gs.CurrentLoop}},
	})

	if gs.CurrentDay > script.GetDaysPerLoop() {
		// End of the loop
		ge.Logger().Info("Loop has ended. Resetting for the next loop.")
		gs.CurrentLoop++
		ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_LOOP_RESET, &model.EventPayload{
			Payload: &model.EventPayload_LoopReset{LoopReset: &model.LoopResetEvent{LoopNumber: gs.CurrentLoop}},
		})

		// Check for game over condition after loop reset
		if gs.CurrentLoop > script.GetLoopCount() {
			ge.Logger().Info("Final loop has ended. Game over.")
			return &GameOverPhase{}
		}

		// Reset for the new loop
		resetForNewLoop(ge)
		return &DayStartPhase{}
	}

	// Reset for the new day
	resetForNewDay(ge)
	return &DayStartPhase{}
}

func (p *IncidentsPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	return nil
}
func (p *IncidentsPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }
func (p *IncidentsPhase) HandleTimeout(ge GameEngine) Phase                     { return nil }
func (p *IncidentsPhase) Exit(ge GameEngine)                                     {}
func (p *IncidentsPhase) Type() model.GamePhase                                  { return model.GamePhase_GAME_PHASE_INCIDENTS }
func (p *IncidentsPhase) TimeoutDuration() time.Duration                         { return 0 }

// resetForNewDay resets the daily state of the game.
func resetForNewDay(ge GameEngine) {
	gs := ge.GetGameState()
	gs.PlayedCardsThisDay = make(map[int32]*model.CardList)
	gs.DayEvents = nil
	// Other daily resets can go here.
}

// resetForNewLoop resets the loop-specific state of the game.
func resetForNewLoop(ge GameEngine) {
	gs := ge.GetGameState()
	gs.CurrentDay = 1
	gs.PlayedCardsThisLoop = make(map[int32]bool)
	gs.TriggeredIncidents = make(map[string]bool)
	gs.LoopEvents = nil

	// Reset characters to their initial state
	script := ge.GetGameRepo().GetScript()
	for _, charInScript := range script.Characters {
		if char, ok := gs.Characters[charInScript.CharacterId]; ok {
			char.CurrentLocation = charInScript.InitialLocation
			char.Paranoia = 0
			char.Intrigue = 0
			char.Goodwill = 0
			char.Traits = nil // Or reset to initial traits
			// Reset other character stats as needed
		}
	}

	resetForNewDay(ge) // Also perform daily reset
}

func init() {
	RegisterPhase(&IncidentsPhase{})
}