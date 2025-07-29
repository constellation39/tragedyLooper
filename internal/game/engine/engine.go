package engine

import (
	"time"
	"tragedylooper/internal/game/engine/phases"
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

	gameConfig loader.GameConfig

	engineChan chan engineRequest
	stopChan   chan struct{}

	dispatchGameEvent chan *model.GameEvent

	currentPhase phases.Phase
	phaseTimer   *time.Timer
	gameStarted  bool

	playerReady map[int32]bool

	mastermindPlayerID   int32
	protagonistPlayerIDs []int32
}

// NewGameEngine creates a new game engine instance.
func NewGameEngine(logger *zap.Logger, players []*model.Player, llmClient llm.Client, gameConfig loader.GameConfig) (*GameEngine, error) {
	ge := &GameEngine{
		logger:               logger,
		llmClient:            llmClient,
		gameConfig:           gameConfig,
		engineChan:           make(chan engineRequest, 100),
		stopChan:             make(chan struct{}),
		dispatchGameEvent:    make(chan *model.GameEvent, 100),
		currentPhase:         &phases.SetupPhase{}, // Start with the new SetupPhase
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
	// Initial transition
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

	// Initial phase enter
	nextPhase := ge.currentPhase.Enter(ge)
	ge.transitionTo(nextPhase)

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
			nextPhase := ge.currentPhase.HandleTimeout(ge)
			ge.transitionTo(nextPhase)
		}
	}
}

func (ge *GameEngine) handleEvent(event *model.GameEvent) {
	nextPhase := ge.currentPhase.HandleEvent(ge, event)
	ge.transitionTo(nextPhase)
}

func (ge *GameEngine) transitionTo(nextPhase phases.Phase) {
	if nextPhase == nil {
		// No transition, stay in the current phase
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

	// Call Enter on the new phase, which may return another phase to transition to immediately.
	followingPhase := ge.currentPhase.Enter(ge)

	duration := ge.currentPhase.TimeoutDuration()
	if duration > 0 {
		ge.phaseTimer.Reset(duration)
	}

	if followingPhase != nil {
		ge.transitionTo(followingPhase)
	}
}

func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.applyAndPublishEvent(model.GameEventType_GAME_ENDED, &model.GameOverEvent{Winner: winner})
	ge.logger.Info("Game over", zap.String("winner", winner.String()))
	ge.transitionTo(&phases.GameOverPhase{})
}

func (ge *GameEngine) ResetPlayerReadiness() {
	for playerID := range ge.GameState.Players {
		ge.playerReady[playerID] = false
	}
}

func (ge *GameEngine) getCharacterByID(charID int32) *model.Character {
	char, ok := ge.GameState.Characters[charID]
	if !ok {
		return nil
	}
	return char
}

func (ge *GameEngine) TriggerIncidents() {
	// TODO: Implement incident triggering logic
}

func (ge *GameEngine) getPlayerByID(playerID int32) *model.Player {
	player, ok := ge.GameState.Players[playerID]
	if !ok {
		return nil
	}
	return player
}

// GetGameState implements the phases.GameEngine interface.
func (ge *GameEngine) GetGameState() *model.GameState {
	return ge.GameState
}

// GetGameConfig implements the phases.GameEngine interface.
func (ge *GameEngine) GetGameConfig() loader.GameConfig {
	return ge.gameConfig
}

func (ge *GameEngine) ApplyAndPublishEvent(eventType model.GameEventType, eventData interface{}) {
	// TODO: implement me
}

func (ge *GameEngine) AreAllPlayersReady() bool {
	// TODO: implement me
	return false
}

func (ge *GameEngine) ResolveMovement() {
	// TODO: implement me
}

func (ge *GameEngine) ResolveOtherCards() {
	// TODO: implement me
}
