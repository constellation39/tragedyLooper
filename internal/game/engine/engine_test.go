package engine

import (
	"testing"
	"time"
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/v1"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockGameDataAccessor provides a mock implementation of the GameConfig interface for testing.

type MockGameDataAccessor struct {
	loader.GameConfig
}

// MockLLMClient provides a mock implementation of the LLM client.
type MockLLMClient struct{}

func (m *MockLLMClient) GenerateResponse(prompt string, sessionID string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func newTestGameEngine(_ *testing.T, logger *zap.Logger, players []*model.Player, data loader.GameConfig) *GameEngine {
	ge, _ := NewGameEngine(logger, players, &MockLLMClient{}, data)
	return ge
}

var (
	testLogger         *zap.Logger
	testPlayers        []*model.Player
	testGameData       loader.GameConfig
	testMastermindOnly []*model.Player
	testEmptyPlayers   []*model.Player
)

func init() {
	var err error
	testLogger, err = zap.NewDevelopment()
	if err != nil {
		panic("Failed to create logger for tests: " + err.Error())
	}

	testPlayers = []*model.Player{
		{Id: 1, Role: model.PlayerRole_MASTERMIND},
		{Id: 2, Role: model.PlayerRole_PROTAGONIST},
		{Id: 3, Role: model.PlayerRole_PROTAGONIST},
	}

	gameLoader := loader.NewJSONLoader("../../../data")
	testGameData, err = gameLoader.LoadGameDataAccessor("first_steps")
	if err != nil {
		testLogger.Fatal("Failed to load script for tests", zap.Error(err))
	}

	testMastermindOnly = []*model.Player{{Id: 1}}
	testEmptyPlayers = []*model.Player{}
}

func TestNewGameEngine(t *testing.T) {
	ge := newTestGameEngine(t, testLogger, testPlayers, testGameData)

	assert.NotNil(t, ge)
	// assert.Equal(t, "test-game", ge.GameState.GameId)
	assert.Equal(t, int32(1), ge.mastermindPlayerID)
	assert.ElementsMatch(t, []int32{2, 3}, ge.protagonistPlayerIDs)
	assert.Len(t, ge.GameState.Characters, 6)
}

func TestGameLoopLifecycle(t *testing.T) {
	ge := newTestGameEngine(t, testLogger, testEmptyPlayers, testGameData)

	ge.StartGameLoop()
	// Give the loop a moment to start
	time.Sleep(20 * time.Millisecond)

	ge.StopGameLoop()

	// Check if the control channel is closed. Reading from a closed channel returns immediately.
	// This confirms the loop has received the stop signal.
	select {
	case _, ok := <-ge.stopChan:
		assert.False(t, ok, "stopChan should be closed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("StopGameLoop did not close the channel within the timeout")
	}
}

func TestSubmitPlayerAction(t *testing.T) {
	ge := newTestGameEngine(t, testLogger, testMastermindOnly, testGameData)
	ge.StartGameLoop()
	defer ge.StopGameLoop()

	action := &model.PlayerActionPayload{Payload: &model.PlayerActionPayload_PlayCard{PlayCard: &model.PlayCardPayload{CardId: 1, Target: &model.PlayCardPayload_TargetCharacterId{TargetCharacterId: 1}}}}
	ge.SubmitPlayerAction(1, action)

	// Check if the action was received by the loop
	select {
	case req := <-ge.engineChan:
		llmReq, ok := req.(*llmActionCompleteRequest)
		assert.True(t, ok)
		assert.Equal(t, int32(1), llmReq.playerID)
		assert.Equal(t, action, llmReq.action)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Action was not received by the game engine")
	}
}

func TestGetPlayerView(t *testing.T) {
	ge := newTestGameEngine(t, testLogger, testPlayers, testGameData)
	ge.StartGameLoop()
	defer ge.StopGameLoop()

	// Allow the game loop to process the request
	time.Sleep(50 * time.Millisecond)

	view := ge.GetPlayerView(1)
	assert.NotNil(t, view, "GetPlayerView should return a non-nil view")
	assert.Equal(t, "test-game", view.GameId)

	// Test that a non-existent player gets a nil view
	view = ge.GetPlayerView(999)
	assert.Nil(t, view, "GetPlayerView for a non-existent player should be nil")
}

func TestPhaseTransitions(t *testing.T) {
	ge := newTestGameEngine(t, testLogger, testPlayers, testGameData)
	ge.StartGameLoop()
	defer ge.StopGameLoop()

	// Initial state should be SETUP
	assert.Equal(t, model.GamePhase_SETUP, ge.GameState.CurrentPhase)

	// Trigger morning phase logic by waiting for the loop to process it
	time.Sleep(150 * time.Millisecond)

	// Should transition to CARD_PLAY
	assert.Equal(t, model.GamePhase_CARD_PLAY, ge.GameState.CurrentPhase)

	// Check if a DAY_ADVANCED event was sent
	select {
	case event := <-ge.dispatchGameEvent:
		assert.Equal(t, model.GameEventType_DAY_ADVANCED, event.Type)
		payload := &model.DayAdvancedEvent{}
		err := event.Payload.UnmarshalTo(payload)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), payload.Day)
		assert.Equal(t, int32(1), payload.Loop)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive DAY_ADVANCED event")
	}
}

func TestCharacterStateChanges(t *testing.T) {
	ge := newTestGameEngine(t, testLogger, testPlayers, testGameData)
	ge.StartGameLoop()
	defer ge.StopGameLoop()

	// Test Paranoia Adjustment
	ge.applyAndPublishEvent(model.GameEventType_PARANOIA_ADJUSTED, &model.ParanoiaAdjustedEvent{CharacterId: 1, NewParanoia: 5, Amount: 5})
	select {
	case event := <-ge.dispatchGameEvent:
		assert.Equal(t, model.GameEventType_PARANOIA_ADJUSTED, event.Type)
		payload := &model.ParanoiaAdjustedEvent{}
		err := event.Payload.UnmarshalTo(payload)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), payload.CharacterId)
		assert.Equal(t, int32(5), payload.NewParanoia)
		assert.Equal(t, int32(5), ge.GameState.Characters[1].Paranoia)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive ParanoiaAdjusted event")
	}

	// Test Goodwill Adjustment
	ge.applyAndPublishEvent(model.GameEventType_GOODWILL_ADJUSTED, &model.GoodwillAdjustedEvent{CharacterId: 1, NewGoodwill: 3, Amount: 3})
	select {
	case event := <-ge.dispatchGameEvent:
		assert.Equal(t, model.GameEventType_GOODWILL_ADJUSTED, event.Type)
		payload := &model.GoodwillAdjustedEvent{}
		err := event.Payload.UnmarshalTo(payload)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), payload.CharacterId)
		assert.Equal(t, int32(3), payload.NewGoodwill)
		assert.Equal(t, int32(3), ge.GameState.Characters[1].Goodwill)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive GoodwillAdjusted event")
	}

	// Test Intrigue Adjustment
	ge.applyAndPublishEvent(model.GameEventType_INTRIGUE_ADJUSTED, &model.IntrigueAdjustedEvent{CharacterId: 1, NewIntrigue: 7, Amount: 7})
	select {
	case event := <-ge.dispatchGameEvent:
		assert.Equal(t, model.GameEventType_INTRIGUE_ADJUSTED, event.Type)
		payload := &model.IntrigueAdjustedEvent{}
		err := event.Payload.UnmarshalTo(payload)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), payload.CharacterId)
		assert.Equal(t, int32(7), payload.NewIntrigue)
		assert.Equal(t, int32(7), ge.GameState.Characters[1].Intrigue)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive IntrigueAdjusted event")
	}

	// Test Location Change
	ge.applyAndPublishEvent(model.GameEventType_CHARACTER_MOVED, &model.CharacterMovedEvent{CharacterId: 1, NewLocation: model.LocationType_SCHOOL})
	select {
	case event := <-ge.dispatchGameEvent:
		assert.Equal(t, model.GameEventType_CHARACTER_MOVED, event.Type)
		payload := &model.CharacterMovedEvent{}
		err := event.Payload.UnmarshalTo(payload)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), payload.CharacterId)
		assert.Equal(t, model.LocationType_SCHOOL, payload.NewLocation)
		assert.Equal(t, model.LocationType_SCHOOL, ge.GameState.Characters[1].CurrentLocation)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive CharacterMoved event")
	}

	select {
	case <-time.After(time.Hour):
	}
}
