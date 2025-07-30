package engine

import (
	"context"
	"testing"
	"tragedylooper/internal/game/engine/phase"
	"tragedylooper/internal/game/loader"
	"tragedylooper/internal/logger"
	v1 "tragedylooper/pkg/proto/v1"

	"github.com/stretchr/testify/assert"
)

// mockActionGenerator is a simple AI action generator for testing purposes.
type mockActionGenerator struct{}

func (m *mockActionGenerator) GenerateAction(ctx context.Context, agc *ActionGeneratorContext) (*v1.PlayerActionPayload, error) {
	// For this test, we don't need complex AI logic.
	// We can return a simple pass action.
	return &v1.PlayerActionPayload{
		Payload: &v1.PlayerActionPayload_PassTurn{
			PassTurn: &v1.PassTurnAction{},
		},
	}, nil
}

func helper_NewGameEngineForTest(t *testing.T) *GameEngine {
	t.Helper()

	log := logger.New()

	gameConfig, err := loader.LoadConfig("../../../data", "first_steps")
	if err != nil {
		t.Fatalf("failed to load game data: %v", err)
	}

	players := []*v1.Player{
		{Id: 1, Name: "Mastermind", Role: v1.PlayerRole_MASTERMIND, IsLlm: true},
		{Id: 2, Name: "Protagonist 1", Role: v1.PlayerRole_PROTAGONIST},
		{Id: 3, Name: "Protagonist 2", Role: v1.PlayerRole_PROTAGONIST},
	}

	engine, err := NewGameEngine(log, players, &mockActionGenerator{}, gameConfig)
	if err != nil {
		t.Fatalf("failed to create game engine: %v", err)
	}

	return engine
}

// TestEngine_Integration_CardPlayAndIncidentTrigger validates the entire data flow:
// 1. load game data from JSON files.
// 2. Initialize the engine.
// 3. A player plays a card to increase a character's paranoia.
// 4. Verify the paranoia stat is updated.
// 5. Continue increasing paranoia until an incident is triggered.
// 6. Verify the incident's effects are applied (a character gains a new trait).
func TestEngine_Integration_CardPlayAndIncidentTrigger(t *testing.T) {
	engine := helper_NewGameEngineForTest(t)
	engine.StartGameLoop()
	defer engine.StopGameLoop()

	// --- Step 1: Initial State Verification ---
	mastermind := engine.getPlayerByID(1)
	assert.NotNil(t, mastermind)

	// The Doctor (Character ID 2) is the Killer in this script.
	doctor := engine.GetCharacterByID(2)
	assert.NotNil(t, doctor)
	assert.Equal(t, int32(0), doctor.Paranoia, "Initial paranoia should be 0")

	// --- Step 2: Mastermind plays a card to increase Paranoia ---
	// Card ID 4 is "Add Paranoia" for the Mastermind
	playCardAction := &v1.PlayerActionPayload{
		Payload: &v1.PlayerActionPayload_PlayCard{
			PlayCard: &v1.PlayCardPayload{
				CardId: 4,
				Target: &v1.PlayCardPayload_TargetCharacterId{TargetCharacterId: doctor.Config.Id},
			},
		},
	}

	// The engine expects the game to be in the Main phase for a card play.
	// We manually set the phase for this test.
	engine.pm.transitionTo(&phase.CardResolvePhase{})

	// Submit the action
	engine.SubmitPlayerAction(mastermind.Id, playCardAction)

	// --- Step 3: Verify the direct effect of the card play ---
	// We need to wait for the engine to process the action.
	// A simple way is to request a player view, which blocks until the engine is free.
	_ = engine.GetPlayerView(mastermind.Id)

	doctorAfterCardPlay := engine.GetCharacterByID(2)
	assert.NotNil(t, doctorAfterCardPlay)
	assert.Equal(t, int32(1), doctorAfterCardPlay.Paranoia, "Paranoia should have increased by 1")

	// --- Step 4: Trigger the "Murder" incident ---
	// The "Murder" incident in "first_steps" requires the Killer (Doctor) to have paranoia >= 3
	// and be at the same location as the Target (High School Girl, ID 1).
	// Let's move the High School Girl to the Hospital.
	highSchoolGirl := engine.GetCharacterByID(1)
	assert.NotNil(t, highSchoolGirl)
	engine.moveCharacter(highSchoolGirl, 0, 1) // Move to Hospital (0,1)
	_ = engine.GetPlayerView(mastermind.Id)    // Sync with engine

	// Now, play the "Add Paranoia" card two more times.
	for i := 0; i < 2; i++ {
		engine.SubmitPlayerAction(mastermind.Id, playCardAction)
		_ = engine.GetPlayerView(mastermind.Id) // Sync with engine
	}

	doctorAfterMultiplePlays := engine.GetCharacterByID(2)
	assert.NotNil(t, doctorAfterMultiplePlays)
	assert.Equal(t, int32(3), doctorAfterMultiplePlays.Paranoia, "Paranoia should be 3")

	// --- Step 5: Verify the incident was triggered automatically ---
	// With Paranoia at 3 and both characters at the Hospital, the "Murder" incident should trigger automatically.
	// The engine's internal `checkForTriggers` should have fired after the last state change.
	doctorAfterIncident := engine.GetCharacterByID(2)
	assert.NotNil(t, doctorAfterIncident)

	foundTrait := false
	for _, trait := range doctorAfterIncident.Traits {
		if trait == "Culprit" {
			foundTrait = true
			break
		}
	}
	assert.True(t, foundTrait, "The 'Culprit' trait should have been added to the Doctor after the incident triggered")

	// --- Step 6: Verify incident triggered event ---
	// The event manager should have broadcasted an IncidentTriggeredEvent
	var triggeredEvent *v1.GameEvent
	for event := range engine.GetGameEvents() {
		if event.Type == v1.GameEventType_INCIDENT_TRIGGERED {
			triggeredEvent = event
			break
		}
	}
	assert.NotNil(t, triggeredEvent, "An IncidentTriggeredEvent should have been published")
	if triggeredEvent != nil {
		payload := triggeredEvent.Payload.GetIncidentTriggered()
		assert.Equal(t, "Murder", payload.Incident)
		assert.Equal(t, doctor.Config.Id, payload.Incident)
	}
}
