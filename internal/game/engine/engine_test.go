package engine

import (
	"context"
	"testing"
	"time"
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
		{Id: 2, Name: "Protagonist 1", Role: v1.PlayerRole_PROTAGONIST, IsLlm: true},
		{Id: 3, Name: "Protagonist 2", Role: v1.PlayerRole_PROTAGONIST, IsLlm: true},
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

	// --- Setup: Pre-position characters and prepare the Mastermind's hand ---
	mastermind := engine.getPlayerByID(1)
	doctor := engine.GetCharacterByID(2)
	highSchoolGirl := engine.GetCharacterByID(1)
	assert.NotNil(t, mastermind)
	assert.NotNil(t, doctor)
	assert.NotNil(t, highSchoolGirl)

	// Move High School Girl to the same location as the Doctor (Hospital)
	doctor.CurrentLocation = highSchoolGirl.CurrentLocation

	// Give the Mastermind three "Add Paranoia" cards (ID 4)
	card4Config, err := loader.Get[*v1.CardConfig](engine.gameConfig, 4)
	assert.NoError(t, err)
	mastermind.Hand = []*v1.Card{
		{Config: card4Config},
		{Config: card4Config},
		{Config: card4Config},
	}

	// --- Execution: Start the game and let it run ---
	engine.StartGameLoop()
	defer engine.StopGameLoop()

	// Mastermind plays all three cards
	for i := 0; i < 3; i++ {
		playCardAction := &v1.PlayerActionPayload{
			Payload: &v1.PlayerActionPayload_PlayCard{
				PlayCard: &v1.PlayCardPayload{
					CardId: 4,
					Target: &v1.PlayCardPayload_TargetCharacterId{TargetCharacterId: doctor.Config.Id},
				},
			},
		}
		engine.SubmitPlayerAction(mastermind.Id, playCardAction)
	}

	// --- Verification: Wait for the day to end and then check the state ---
	waitForEvent(t, engine, v1.GameEventType_DAY_ADVANCED)

	doctorAfterIncident := engine.GetCharacterByID(2)
	assert.NotNil(t, doctorAfterIncident)
	assert.Equal(t, int32(3), doctorAfterIncident.Paranoia, "Paranoia should be 3")

	foundTrait := false
	for _, trait := range doctorAfterIncident.Traits {
		if trait == "Culprit" {
			foundTrait = true
			break
		}
	}
	assert.True(t, foundTrait, "The 'Culprit' trait should have been added to the Doctor")
}

// waitForEvent is a helper function to block until the game engine emits a specific event.
func waitForEvent(t *testing.T, engine *GameEngine, targetEvent v1.GameEventType) {
	t.Helper()
	timeout := time.After(2 * time.Second) // 2-second timeout

	// We need to consume events from the channel to find our target event.
	// This should be done in a separate goroutine to avoid blocking the main test thread
	// if the event never comes.
	eventChan := engine.GetGameEvents()

	for {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for event %s", targetEvent)
		case event := <-eventChan:
			if event.Type == targetEvent {
				return
			}
		}
	}
}
