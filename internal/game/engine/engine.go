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
	internalEventChan chan *model.GameEvent // New channel for internal events
	stopChan          chan struct{}
	llmClient         llm.Client
	playerReady       map[int32]bool
	logger            *zap.Logger
	currentPhase      Phase
	phaseTimer        *time.Timer
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
	initialPhase := phaseImplementations[model.GamePhase_MASTERMIND_SETUP]
	gs := &model.GameState{
		GameId:                  "",
		Characters:              make(map[int32]*model.Character),
		Players:                 players,
		CurrentDay:              1,
		CurrentLoop:             1,
		CurrentPhase:            initialPhase.Type(),
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
		internalEventChan: make(chan *model.GameEvent, 100),
		stopChan:          make(chan struct{}),
		llmClient:         llmClient,
		playerReady:       make(map[int32]bool),
		logger:            logger,
		currentPhase:      initialPhase,
	}

	ge.dealInitialCards()

	return ge, nil
}

func (ge *GameEngine) StartGameLoop() {
	go ge.runGameLoop()
	// Start the first phase
	ge.transitionTo(ge.currentPhase)
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

	// Initial timer setup
	ge.phaseTimer = time.NewTimer(time.Hour) // Start with a long duration, will be reset
	ge.phaseTimer.Stop()

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

		case event := <-ge.internalEventChan:
			ge.handleEvent(event)

		case <-ge.phaseTimer.C:
			ge.handleTimeout()
		}
	}
}

func (ge *GameEngine) handleEvent(event *model.GameEvent) {
	nextPhase := ge.currentPhase.HandleEvent(ge, event)
	if nextPhase != nil {
		ge.transitionTo(nextPhase)
	}
}

func (ge *GameEngine) handleTimeout() {
	nextPhase := ge.currentPhase.HandleTimeout(ge)
	if nextPhase != nil {
		ge.transitionTo(nextPhase)
	}
}

func (ge *GameEngine) transitionTo(nextPhase Phase) {
	ge.phaseTimer.Stop()

	ge.logger.Info("Transitioning phase", zap.String("from", ge.currentPhase.Type().String()), zap.String("to", nextPhase.Type().String()))
	ge.currentPhase.Exit(ge)
	ge.currentPhase = nextPhase
	ge.GameState.CurrentPhase = nextPhase.Type()
	ge.currentPhase.Enter(ge)

	duration := ge.currentPhase.TimeoutDuration()
	if duration > 0 {
		ge.phaseTimer.Reset(duration)
	}

	// Some phases transition automatically, so we send a nil event to trigger the next transition.
	ge.internalEventChan <- &model.GameEvent{Type: model.GameEventType_GAME_EVENT_TYPE_UNSPECIFIED}
}

func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.applyAndPublishEvent(model.GameEventType_GAME_OVER, &model.GameOverEvent{Winner: winner})
	ge.logger.Info("Game over", zap.String("winner", winner.String()))
	ge.transitionTo(phaseImplementations[model.GamePhase_GAME_OVER])
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