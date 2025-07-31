package engine

import (
	"context"
	"testing"
	"time"
	"tragedylooper/internal/game/engine/ai"
	"tragedylooper/internal/game/loader"
	"tragedylooper/internal/logger"
	v1 "tragedylooper/pkg/proto/tragedylooper/v1"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// smartMockActionGenerator is a more intelligent mock for the AI player.
// It will attempt to play cards that are relevant to the current game state.
type smartMockActionGenerator struct {
	t      *testing.T
	logger *zap.Logger
}

func (m *smartMockActionGenerator) GenerateAction(ctx context.Context, agc *ai.ActionGeneratorContext) (*v1.PlayerActionPayload, error) {
	m.logger.Info("Smart mock AI generating action for player", zap.Int32("playerID", agc.Player.Id))

	doctor := helper_GetCharacterFromView(m.t, agc.PlayerView, 2)

	// Mastermind logic: always try to increase the Doctor's paranoia.
	if agc.Player.Role == v1.PlayerRole_PLAYER_ROLE_MASTERMIND {
		for _, card := range agc.Player.Hand.Cards {
			if card.Config.Id == 4 { // "Increase Paranoia" card
				m.logger.Info("Mastermind playing 'Increase Paranoia'", zap.Int32("targetID", doctor.Id))
				return &v1.PlayerActionPayload{
					Payload: &v1.PlayerActionPayload_PlayCard{
						PlayCard: &v1.PlayCardPayload{
							CardId: card.Config.Id,
							Target: &v1.PlayCardPayload_TargetCharacterId{TargetCharacterId: doctor.Id},
						},
					},
				}, nil
			}
		}
	}

	// Protagonist logic: always try to decrease the Doctor's paranoia.
	if agc.Player.Role == v1.PlayerRole_PLAYER_ROLE_PROTAGONIST {
		for _, card := range agc.Player.Hand.Cards {
			if card.Config.Id == 2 { // "Decrease Paranoia" card
				m.logger.Info("Protagonist playing 'Decrease Paranoia'", zap.Int32("targetID", doctor.Id))
				return &v1.PlayerActionPayload{
					Payload: &v1.PlayerActionPayload_PlayCard{
						PlayCard: &v1.PlayCardPayload{
							CardId: card.Config.Id,
							Target: &v1.PlayCardPayload_TargetCharacterId{TargetCharacterId: doctor.Id},
						},
					},
				}, nil
			}
		}
	}

	// Default action: pass
	m.logger.Info("AI passing turn", zap.Int32("playerID", agc.Player.Id))
	return &v1.PlayerActionPayload{
		Payload: &v1.PlayerActionPayload_PassTurn{
			PassTurn: &v1.PassTurnAction{},
		},
	}, nil
}

func helper_NewGameEngineForFullGameTest(t *testing.T) *GameEngine {
	t.Helper()

	log := logger.New()

	gameConfig, err := loader.LoadConfig("../../../data", "first_steps")
	if err != nil {
		t.Fatalf("failed to load game data: %v", err)
	}

	players := []*v1.Player{
		{Id: 1, Name: "Mastermind", Role: v1.PlayerRole_PLAYER_ROLE_MASTERMIND, IsLlm: true},
		{Id: 2, Name: "Protagonist 1", Role: v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, IsLlm: true},
		{Id: 3, Name: "Protagonist 2", Role: v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, IsLlm: true},
		{Id: 4, Name: "Protagonist 3", Role: v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, IsLlm: true},
	}

	actionGenerator := &smartMockActionGenerator{t: t, logger: log}
	engine, err := NewGameEngine(log, players, actionGenerator, gameConfig)
	if err != nil {
		t.Fatalf("failed to create game engine: %v", err)
	}

	return engine
}

// TestEngine_FullGame_ProtagonistWin simulates a full game where the protagonists win.
func TestEngine_FullGame_ProtagonistWin(t *testing.T) {
	engine := helper_NewGameEngineForFullGameTest(t)
	engine.Start()
	defer engine.Stop()

	// We expect the game to end. The exact number of events is hard to predict,
	// so we'll just wait for the game over event.
	testCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gameOver := false
	eventChan := engine.GetGameEvents()

	for !gameOver {
		select {
		case event := <-eventChan:
			t.Logf("Received event: %s", event.Type)
			if event.Type == v1.GameEventType_GAME_EVENT_TYPE_GAME_ENDED {
				gameOver = true
				payload := event.GetPayload().GetGameEnded()
				assert.NotNil(t, payload)
				assert.Equal(t, v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, payload.Winner, "Protagonists should win")
			}
		case <-testCtx.Done():
			t.Fatal("Test timed out waiting for game to end")
		}
	}

	assert.True(t, gameOver, "Game should have ended")
}
