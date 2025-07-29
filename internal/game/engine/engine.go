package engine

import (
	"context"
	"tragedylooper/internal/game/engine/eventhandler"
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GameEngine manages the state and logic of a single game instance.
type engineRequest interface{}

// getPlayerViewRequest is a request to get a filtered view of the game state for a player.
type getPlayerViewRequest struct {
	playerID     int32
	responseChan chan *model.PlayerView
}

type aiActionCompleteRequest struct {
	playerID int32
	action   *model.PlayerActionPayload
}
type GameEngine struct {
	GameState *model.GameState
	logger    *zap.Logger

	actionGenerator ActionGenerator
	gameConfig      loader.GameConfig
	pm              *phaseManager

	// engineChan is the central channel for all incoming requests (player actions, AI actions, etc.).
	// It ensures that all modifications to the game state are processed sequentially in the main game loop,
	// preventing race conditions.
	engineChan chan engineRequest
	stopChan   chan struct{}

	// dispatchGameEvent is an outbound channel that broadcasts processed game events to external listeners.
	dispatchGameEvent chan *model.GameEvent

	playerReady map[int32]bool

	mastermindPlayerID   int32
	protagonistPlayerIDs []int32
}

// NewGameEngine creates a new game engine instance.
func NewGameEngine(logger *zap.Logger, players []*model.Player, actionGenerator ActionGenerator, gameConfig loader.GameConfig) (*GameEngine, error) {
	ge := &GameEngine{
		logger:               logger,
		actionGenerator:      actionGenerator,
		gameConfig:           gameConfig,
		engineChan:           make(chan engineRequest, 100),
		stopChan:             make(chan struct{}),
		dispatchGameEvent:    make(chan *model.GameEvent, 100),
		playerReady:          make(map[int32]bool),
		mastermindPlayerID:   0,
		protagonistPlayerIDs: nil,
	}
	ge.pm = newPhaseManager(ge)

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
	case ge.engineChan <- &aiActionCompleteRequest{playerID: playerID, action: action}:
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

	// This blocks until the main game loop processes the request and sends a response.
	ge.engineChan <- req
	view := <-responseChan
	return view
}

// runGameLoop is the heart of the game engine. It's a single-threaded loop that processes all game events
// and state changes sequentially, ensuring thread safety without complex locking.
func (ge *GameEngine) runGameLoop() {
	ge.logger.Info("Game loop started.")
	defer ge.logger.Info("Game loop stopped.")

	// The phase manager is started, which initiates the first phase transition.
	ge.pm.start()
	defer ge.pm.stop()

	for {
		select {
		case <-ge.stopChan:
			return
		case req := <-ge.engineChan:
			ge.handleEngineRequest(req)
		case <-ge.pm.timer():
			ge.handleTimeout()
		}
	}
}

// handleEngineRequest processes incoming requests from the engine's channel.
func (ge *GameEngine) handleEngineRequest(req engineRequest) {
	switch r := req.(type) {
	case *aiActionCompleteRequest:
		// An AI or player has submitted an action.
		ge.pm.handleAction(r.playerID, r.action)
	case *getPlayerViewRequest:
		// A request for a player-specific view of the game state.
		r.responseChan <- ge.GeneratePlayerView(r.playerID)
	default:
		ge.logger.Warn("Unhandled request type in engine channel")
	}
}

// handleTimeout is called when the current phase's timer expires.
func (ge *GameEngine) handleTimeout() {
	ge.pm.handleTimeout()
}

// CreateAndProcessEvent is the central method for creating, applying, and broadcasting game events.
// It ensures a consistent order of operations:
// 1. The event is created from a payload.
// 2. The game state is mutated synchronously by the event handler.
// 3. The current phase is allowed to react to the event, potentially triggering a phase transition.
// 4. The event is broadcast to external listeners and recorded in the game's history.
func (ge *GameEngine) CreateAndProcessEvent(eventType model.GameEventType, payload proto.Message) {
	anyPayload, err := anypb.New(payload)
	if err != nil {
		ge.logger.Error("Failed to create anypb.Any for event payload", zap.Error(err))
		return
	}
	event := &model.GameEvent{
		Type:      eventType,
		Payload:   anyPayload,
		Timestamp: timestamppb.Now(),
	}

	// Step 1: Apply the state change synchronously.
	// This is critical to ensure the game state is consistent before any other logic runs.
	if err := eventhandler.ProcessEvent(ge.GameState, event); err != nil {
		ge.logger.Error("Failed to apply event to game state", zap.Error(err), zap.String("type", event.Type.String()))
		// We continue even if the handler fails, to allow the phase and listeners to react.
	}

	// Step 2: Let the current phase react to the event.
	// This is now handled by the phase manager.
	ge.pm.handleEvent(event)

	// Step 3: Publish the event to external listeners and record it.
	// This happens after the state has been updated.
	select {
	case ge.dispatchGameEvent <- event:
		// Also record the event in the game state for player views
		ge.GameState.DayEvents = append(ge.GameState.DayEvents, event)
		ge.GameState.LoopEvents = append(ge.GameState.LoopEvents, event)
	default:
		ge.logger.Warn("Game event channel full, dropping event", zap.String("eventType", event.Type.String()))
	}

	// TODO: Re-implement trigger logic here, after state has fully updated.
	// ge.checkForTriggers(event)
}

func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.logger.Info("Game over", zap.String("winner", winner.String()))
	// This event will be processed, leading to a state update and a phase transition
	// handled by the current phase's HandleEvent method.
	ge.CreateAndProcessEvent(model.GameEventType_GAME_ENDED, &model.GameOverEvent{Winner: winner})
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

func (ge *GameEngine) MoveCharacter(char *model.Character, dx, dy int) {
	ge.moveCharacter(char, dx, dy)
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

func (ge *GameEngine) GetGameConfig() loader.GameConfig {
	return ge.gameConfig
}

func (ge *GameEngine) AreAllPlayersReady() bool {
	// TODO: implement me
	return false
}

func (ge *GameEngine) Logger() *zap.Logger {
	return ge.logger
}

func (ge *GameEngine) SetPlayerReady(playerID int32) {
	ge.playerReady[playerID] = true
}

// --- AI Integration ---

// TriggerAIPlayerAction prompts an AI player to make a decision.
func (ge *GameEngine) TriggerAIPlayerAction(playerID int32) {
	player := ge.getPlayerByID(playerID)
	if player == nil || !player.IsLlm { // TODO: Make this check more generic (e.g., IsAI)
		return
	}

	ge.logger.Info("Triggering AI for player", zap.String("player", player.Name))

	// Create the context for the action generator
	ctx := &ActionGeneratorContext{
		Player:        player,
		PlayerView:    ge.GetPlayerView(playerID),
		Script:        ge.gameConfig.GetScript(),
		AllCharacters: ge.GameState.Characters,
	}

	go func() {
		action, err := ge.actionGenerator.GenerateAction(context.Background(), ctx)
		if err != nil {
			ge.logger.Error("AI action generation failed", zap.String("player", player.Name), zap.Error(err))
			// Submit a default action to unblock the game
			ge.engineChan <- &aiActionCompleteRequest{
				playerID: playerID,
				action:   &model.PlayerActionPayload{},
			}
			return
		}

		// Send the validated action back to the main loop for processing.
		ge.engineChan <- &aiActionCompleteRequest{
			playerID: playerID,
			action:   action,
		}
	}()
}
