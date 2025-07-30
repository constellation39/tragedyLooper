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
	engine.TriggerIncidents() // Manually trigger incidents for testing
	waitForEvent(t, engine, v1.GameEventType_INCIDENT_TRIGGERED) // Wait for the incident to be processed

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

func TestEngine_GetPlayerView(t *testing.T) {
	engine := helper_NewGameEngineForTest(t)
	engine.StartGameLoop()
	defer engine.StopGameLoop()

	// --- Get Views for Mastermind and a Protagonist ---
	mastermindView := engine.GetPlayerView(1) // Mastermind ID
	protagonistView := engine.GetPlayerView(2) // Protagonist ID

	assert.NotNil(t, mastermindView)
	assert.NotNil(t, protagonistView)

	// --- Verification for Mastermind ---
	// The Mastermind should see the hidden roles of all characters.
	// In 'first_steps', the Doctor (ID 2) is the 'Serial Killer'.
	doctorForMastermind := helper_GetCharacterFromView(t, mastermindView, 2)
	assert.NotNil(t, doctorForMastermind)
	assert.Equal(t, v1.RoleType_KILLER, doctorForMastermind.Role, "Mastermind should see the Doctor's hidden role")

	// --- Verification for Protagonist ---
	// The Protagonist should NOT see the hidden roles. It should be UNKNOWN.
	doctorForProtagonist := helper_GetCharacterFromView(t, protagonistView, 2)
	assert.NotNil(t, doctorForProtagonist)
	assert.Equal(t, v1.RoleType_ROLE_UNKNOWN, doctorForProtagonist.Role, "Protagonist should not see the Doctor's hidden role")

	// Both players should see public information, like the character's name.
	assert.Equal(t, "Doctor", doctorForMastermind.Name)
	assert.Equal(t, "Doctor", doctorForProtagonist.Name)
}

func TestEngine_CharacterMovement(t *testing.T) {
	engine := helper_NewGameEngineForTest(t)
	char := engine.GetCharacterByID(1) // High School Girl, starts at SCHOOL
	assert.NotNil(t, char)
	assert.Equal(t, v1.LocationType_SCHOOL, char.CurrentLocation)

	// --- Test Valid Moves ---
	// Move Right from SCHOOL (1,0) to SHRINE (0,0) - This is horizontal move, dx= -1, dy=0, but the locations are not adjacent this way.
	// Let's check the grid. School is (1,0), Shrine is (0,0). So dx should be -1.
	// Let's move from School (1,0) to City (1,1), dy=1
	engine.MoveCharacter(char, 0, 1)
	assert.Equal(t, v1.LocationType_CITY, char.CurrentLocation, "Should move from School to City")

	// Move Left from CITY (1,1) to HOSPITAL (0,1), dx=-1
	engine.MoveCharacter(char, -1, 0)
	assert.Equal(t, v1.LocationType_HOSPITAL, char.CurrentLocation, "Should move from City to Hospital")

	// --- Test Invalid Moves (Out of Bounds) ---
	// Try to move Left from HOSPITAL (0,1) which is out of bounds
	engine.MoveCharacter(char, -1, 0)
	assert.Equal(t, v1.LocationType_HOSPITAL, char.CurrentLocation, "Should not move out of bounds (left)")

	// Try to move Up from HOSPITAL (0,1) to (0,0) which is Shrine
	engine.MoveCharacter(char, 0, -1)
	assert.Equal(t, v1.LocationType_SHRINE, char.CurrentLocation, "Should move from Hospital to Shrine")

	// Try to move Up from SHRINE (0,0) which is out of bounds
	engine.MoveCharacter(char, 0, -1)
	assert.Equal(t, v1.LocationType_SHRINE, char.CurrentLocation, "Should not move out of bounds (up)")
}

func TestEngine_GameOverOnMaxLoops(t *testing.T) {
	engine := helper_NewGameEngineForTest(t)

	// --- Setup: Give mastermind cards to play ---
	mastermind := engine.GetMastermindPlayer()
	card4Config, err := loader.Get[*v1.CardConfig](engine.gameConfig, 4) // Add Paranoia card
	assert.NoError(t, err)
	mastermind.Hand = []*v1.Card{
		{Config: card4Config},
		{Config: card4Config},
		{Config: card4Config},
	}

	engine.StartGameLoop()
	defer engine.StopGameLoop()

	// Manually set the loop count to the maximum
	engine.GameState.CurrentLoop = engine.gameConfig.GetScript().LoopCount
	// Set day to the last day of the loop
	engine.GameState.CurrentDay = engine.gameConfig.GetScript().DaysPerLoop

	// Wait for the game to be ready for player actions
	waitForPhase(t, engine, v1.GamePhase_CARD_PLAY)

	// --- Mastermind Plays Cards ---
	playCardAction := &v1.PlayerActionPayload{
		Payload: &v1.PlayerActionPayload_PlayCard{
			PlayCard: &v1.PlayCardPayload{
				CardId: 4,
				Target: &v1.PlayCardPayload_TargetCharacterId{TargetCharacterId: 1},
			},
		},
	}
	for i := 0; i < 3; i++ {
		engine.SubmitPlayerAction(mastermind.Id, playCardAction)
	}

	// --- Protagonists Pass ---
	passAction := &v1.PlayerActionPayload{
		Payload: &v1.PlayerActionPayload_PassTurn{
			PassTurn: &v1.PassTurnAction{},
		},
	}
	protagonists := engine.GetProtagonistPlayers()
	for _, player := range protagonists {
		engine.SubmitPlayerAction(player.Id, passAction)
	}

	// We expect a GAME_ENDED event with the protagonist as the winner
	// because the mastermind failed to achieve their goals within the loops.
	waitForEvent(t, engine, v1.GameEventType_GAME_ENDED)

	// We can also check the final phase if needed, but the event is a stronger signal.
	assert.Equal(t, v1.GamePhase_GAME_OVER, engine.GetCurrentPhase(), "Game should be in the GAME_OVER phase")
}

// helper_GetCharacterFromView is a test helper to find a character in a player view by its ID.
func helper_GetCharacterFromView(t *testing.T, view *v1.PlayerView, charID int32) *v1.PlayerViewCharacter {
	t.Helper()
	for _, char := range view.Characters {
		if char.Id == charID {
			return char
		}
	}
	t.Fatalf("Character with ID %d not found in player view", charID)
	return nil
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

// waitForPhase is a helper function to block until the game engine reaches a specific phase.
func waitForPhase(t *testing.T, engine *GameEngine, targetPhase v1.GamePhase) {
	t.Helper()
	timeout := time.After(2 * time.Second)

	for {
		currentPhase := engine.GetCurrentPhase()
		if currentPhase == targetPhase {
			return
		}
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for phase %s, current phase is %s", targetPhase, currentPhase)
		case <-time.After(10 * time.Millisecond):
			// continue polling
		}
	}
}
