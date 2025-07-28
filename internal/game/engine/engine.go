package engine

import (
	"time"
	"tragedylooper/internal/game/loader"

	model "tragedylooper/internal/game/proto/v1"
	"tragedylooper/internal/llm"

	"go.uber.org/zap"
)

// GameEngine manages the state and logic of a single game instance.
type engineRequest interface{}

// getPlayerViewRequest is a request to get a filtered view of the game state for a player.
type getPlayerViewRequest struct {
	playerID     int32
	responseChan chan *model.PlayerView
}

type llmActionCompleteRequest struct {
	playerID int32
	action   *model.PlayerActionPayload
}
type GameEngine struct {
	GameState *model.GameState
	logger    *zap.Logger

	llmClient llm.Client

	gameConfig loader.GameConfigAccessor

	engineChan chan engineRequest
	stopChan   chan struct{}

	dispatchGameEvent chan *model.GameEvent

	currentPhase Phase
	phaseTimer   *time.Timer
	gameStarted  bool

	playerReady map[int32]bool

	mastermindPlayerID   int32
	protagonistPlayerIDs []int32
}

// NewGameEngine creates a new game engine instance.
func NewGameEngine(logger *zap.Logger, players []*model.Player, llmClient llm.Client, gameConfig loader.GameConfigAccessor) (*GameEngine, error) {
	ge := &GameEngine{
		logger:               logger,
		llmClient:            llmClient,
		gameConfig:           gameConfig,
		engineChan:           make(chan engineRequest, 100),
		stopChan:             make(chan struct{}),
		dispatchGameEvent:    make(chan *model.GameEvent, 100),
		currentPhase:         phaseImplementations[model.GamePhase_SETUP],
		phaseTimer:           time.NewTimer(time.Hour),
		playerReady:          make(map[int32]bool),
		mastermindPlayerID:   0,
		protagonistPlayerIDs: nil,
	}

	playerMap := make(map[int32]*model.Player)
	for _, player := range players {
		switch player.Role {
		case model.PlayerRole_MASTERMIND:
			ge.mastermindPlayerID = player.Id
		case model.PlayerRole_PROTAGONIST:
			ge.protagonistPlayerIDs = append(ge.protagonistPlayerIDs, player.Id)
		default:
			ge.logger.Warn("Unknown player role", zap.Int32("playerID", player.Id))
		}

		playerMap[player.Id] = player
	}

	ge.initializeGameStateFromScript(gameConfig, playerMap)
	ge.dealInitialCards()

	return ge, nil
}

func (ge *GameEngine) StartGameLoop() {
	go ge.runGameLoop()
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
	case ge.engineChan <- &llmActionCompleteRequest{playerID: playerID, action: action}:
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

	ge.engineChan <- req
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

		case req := <-ge.engineChan:
			switch in := req.(type) {
			case *llmActionCompleteRequest:
				ge.handlePlayerAction(in.playerID, in.action)
			case *getPlayerViewRequest:
				in.responseChan <- ge.GeneratePlayerView(in.playerID)
			}

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
	if nextPhase == nil {
		ge.logger.Debug("transitionTo called with nil nextPhase, staying in current phase", zap.String("current", ge.currentPhase.Type().String()))
		return
	}

	ge.phaseTimer.Stop()

	if ge.gameStarted {
		ge.logger.Info("Transitioning phase", zap.String("from", ge.currentPhase.Type().String()), zap.String("to", nextPhase.Type().String()))
		ge.currentPhase.Exit(ge)
	} else {
		ge.logger.Info("Entering initial phase", zap.String("to", nextPhase.Type().String()))
		ge.gameStarted = true
	}

	ge.currentPhase = nextPhase
	ge.GameState.CurrentPhase = nextPhase.Type()
	ge.currentPhase.Enter(ge)

	duration := ge.currentPhase.TimeoutDuration()
	if duration > 0 {
		ge.phaseTimer.Reset(duration)
	}
}

func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.applyAndPublishEvent(model.GameEventType_GAME_ENDED, &model.GameOverEvent{Winner: winner})
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
