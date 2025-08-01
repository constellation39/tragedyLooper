package engine

import (
	"context"
	"testing"
	"time"

	"github.com/constellation39/tragedyLooper/internal/game/loader"
	"github.com/constellation39/tragedyLooper/internal/logger"
	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
	"github.com/stretchr/testify/assert"
)

func helper_NewGameEngineForFullGameTest(t *testing.T) *GameEngine {
	t.Helper()

	log := logger.New()

	gameConfig, err := loader.LoadConfig("../../../data", "first_steps")
	if err != nil {
		t.Fatalf("failed to load game data: %v", err)
	}

	players := []*v1.Player{
		{Id: 1, Name: "Mastermind", Role: v1.PlayerRole_PLAYER_ROLE_MASTERMIND, IsLlm: false},
		{Id: 2, Name: "Protagonist 1", Role: v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, IsLlm: false},
		{Id: 3, Name: "Protagonist 2", Role: v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, IsLlm: false},
		{Id: 4, Name: "Protagonist 3", Role: v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, IsLlm: false},
	}

	engine, err := NewGameEngine(log, players, nil, gameConfig)
	if err != nil {
		t.Fatalf("failed to create game engine: %v", err)
	}

	return engine
}

// TestEngine_FullGame_ProtagonistWin simulates a full game where the protagonists win
// by surviving all the loops without the tragedy occurring.
func TestEngine_FullGame_ProtagonistWin(t *testing.T) {
	engine := helper_NewGameEngineForFullGameTest(t)
	engine.Start()
	defer engine.Stop()

	testCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	passAction := &v1.PlayerActionPayload{
		Payload: &v1.PlayerActionPayload_PassTurn{PassTurn: &v1.PassTurnAction{}},
	}

	// Goroutine to automatically pass turns for all players
	go func() {
		for {
			// A small sleep to prevent this from being a tight loop.
			time.Sleep(10 * time.Millisecond)
			select {
			case <-testCtx.Done():
				return
			default:
				phase := engine.GetCurrentPhase()
				if phase == v1.GamePhase_GAME_PHASE_MASTERMIND_CARD_PLAY {
					engine.SubmitPlayerAction(1, passAction)
				} else if phase == v1.GamePhase_GAME_PHASE_PROTAGONIST_CARD_PLAY {
					for _, p := range engine.GetProtagonistPlayers() {
						engine.SubmitPlayerAction(p.Id, passAction)
					}
				}
			}
		}
	}()

	// --- Verification ---
	// Wait for the game to end
	gameOver := false
	eventChan := engine.GetGameEvents()

	for !gameOver {
		select {
		case event := <-eventChan:
			if event.Type == v1.GameEventType_GAME_EVENT_TYPE_GAME_ENDED {
				gameOver = true
				payload := event.GetPayload().GetGameEnded()
				assert.NotNil(t, payload)
				assert.Equal(t, v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, payload.Winner, "Protagonists should win by surviving")
			}
		case <-testCtx.Done():
			t.Fatal("Test timed out waiting for game to end")
		}
	}

	assert.True(t, gameOver, "Game should have ended")
	assert.Equal(t, v1.GamePhase_GAME_PHASE_GAME_OVER, engine.GetCurrentPhase(), "Game should be in the GAME_OVER phase")
}