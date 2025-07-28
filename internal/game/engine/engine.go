package engine

import (
	"time"
	"tragedylooper/internal/game/loader"

	model "tragedylooper/internal/game/proto/v1"
	"tragedylooper/internal/llm"

	"go.uber.org/zap"
)

// GameEngine manages the state and logic of a single game instance.
type GameEngine struct {
	GameState         *model.GameState
	gameConfig        loader.GameConfigAccessor
	requestChan       chan engineRequest
	dispatchGameEvent chan *model.GameEvent
	stopChan          chan struct{}
	llmClient         llm.Client
	playerReady       map[int32]bool
	logger            *zap.Logger
}

// engineRequest is an interface for all requests handled by the game engine loop.
type engineRequest interface{}

// getPlayerViewRequest is a request to get a filtered view of the game state for a player.
type getPlayerViewRequest struct {
	playerID     int32
	responseChan chan *model.PlayerView
}

// llmActionCompleteRequest is sent when an LLM player has decided on an action.
type llmActionCompleteRequest struct {
	playerID int32
	action   *model.PlayerActionPayload
}

// NewGameEngine creates a new game engine instance.
func NewGameEngine(logger *zap.Logger, players map[int32]*model.Player, llmClient llm.Client, gameConfig loader.GameConfigAccessor) (*GameEngine, error) {
	gs := &model.GameState{
		GameId:                  "",
		Characters:              make(map[int32]*model.Character),
		Players:                 players,
		CurrentDay:              1,
		CurrentLoop:             1,
		CurrentPhase:            model.GamePhase_MASTERMIND_SETUP, // Start with Mastermind Setup
		ActiveTragedies:         nil,
		PreventedTragedies:      nil,
		PlayedCardsThisDay:      make(map[int32]*model.Card),
		PlayedCardsThisLoop:     make(map[int32]bool),
		LastUpdateTime:          time.Now().Unix(),
		DayEvents:               []*model.GameEvent{},
		LoopEvents:              []*model.GameEvent{},
		CharacterParanoiaLimits: nil,
		CharacterGoodwillLimits: nil,
		CharacterIntrigueLimits: nil,
	}

	ge := &GameEngine{
		GameState:         gs,
		gameConfig:        gameConfig,
		requestChan:       make(chan engineRequest, 100),
		dispatchGameEvent: make(chan *model.GameEvent, 100),
		stopChan:          make(chan struct{}),
		llmClient:         llmClient,
		playerReady:       make(map[int32]bool),
		logger:            logger,
	}

	ge.dealInitialCards()

	return ge, nil
}

func (ge *GameEngine) StartGameLoop() {
	go ge.runGameLoop()
}

func (ge *GameEngine) StopGameLoop() {
	close(ge.stopChan)
}

func (ge *GameEngine) SubmitPlayerAction(playerID int32, action *model.PlayerActionPayload) {
	if action == nil {
		ge.logger.Warn("Received nil action from player")
		return
	}
	select {
	case ge.requestChan <- &llmActionCompleteRequest{playerID: playerID, action: action}:
	default:
		ge.logger.Warn("Request channel full, dropping action", zap.Int32("playerID", playerID))
	}
}

func (ge *GameEngine) GetGameEvents() <-chan *model.GameEvent {
	return ge.dispatchGameEvent
}

func (ge *GameEngine) GetPlayerView(playerID int32) *model.PlayerView {
	responseChan := make(chan *model.PlayerView)
	req := &getPlayerViewRequest{
		playerID:     playerID,
		responseChan: responseChan,
	}

	ge.requestChan <- req
	view := <-responseChan
	return view
}

func (ge *GameEngine) runGameLoop() {
	ge.logger.Info("Game loop started.")
	defer ge.logger.Info("Game loop stopped.")

	timer := time.NewTicker(100 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ge.stopChan:
			return

		case req := <-ge.requestChan:
			switch r := req.(type) {
			case *llmActionCompleteRequest:
				ge.handlePlayerAction(r.playerID, r.action)
			case *getPlayerViewRequest:
				r.responseChan <- ge.GeneratePlayerView(r.playerID)
			}

		case <-timer.C:
			if handler, ok := phaseHandlers[ge.GameState.CurrentPhase]; ok {
				handler(ge)
			}
		}
	}
}

func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.GameState.CurrentPhase = model.GamePhase_GAME_OVER
	ge.applyAndPublishEvent(model.GameEventType_GAME_OVER, &model.GameOverEvent{Winner: winner})
	ge.logger.Info("Game over", zap.String("winner", winner.String()))
}

func (ge *GameEngine) resetPlayerReadiness() {
	for playerID := range ge.GameState.Players {
		ge.playerReady[playerID] = false
	}
}

func (ge *GameEngine) resetLoop() {
	// Reset character stats
	for _, char := range ge.GameState.Characters {
		char.Paranoia = 0
		char.Goodwill = 0
		char.Intrigue = 0
		// Note: Traits and abilities might persist or reset based on game rules.
		// Current implementation assumes they persist unless explicitly removed.
	}

	// Reset player hands
	for _, p := range ge.GameState.Players {
		p.Hand = nil
	}

	// Reset loop-specific state
	ge.GameState.PlayedCardsThisLoop = make(map[int32]bool)
	ge.GameState.DayEvents = []*model.GameEvent{}
	ge.GameState.LoopEvents = []*model.GameEvent{}

	// Re-deal cards to players
	ge.dealInitialCards()
}

func (ge *GameEngine) findPlayer(playerID int32) *model.Player {
	if p, ok := ge.GameState.Players[playerID]; ok {
		return p
	}
	return nil
}
