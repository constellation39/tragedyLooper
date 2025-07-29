package engine

import (
	"context"
	"time"
	"tragedylooper/internal/game/engine/eventhandler"
	"tragedylooper/internal/game/engine/phase"
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// eventHandlers maps event types to their respective handler instances.
var eventHandlers = map[model.GameEventType]eventhandler.EventHandler{
	model.GameEventType_CHARACTER_MOVED:    &handlers.CharacterMovedHandler{},
	model.GameEventType_PARANOIA_ADJUSTED:  &handlers.ParanoiaAdjustedHandler{},
	model.GameEventType_GOODWILL_ADJUSTED:  &handlers.GoodwillAdjustedHandler{},
	model.GameEventType_INTRIGUE_ADJUSTED:  &eventhandler.IntrigueAdjustedHandler{},
	model.GameEventType_TRAIT_ADDED:        &eventhandler.TraitAddedHandler{},
	model.GameEventType_TRAIT_REMOVED:      &eventhandler.TraitRemovedHandler{},
	model.GameEventType_CARD_PLAYED:        &eventhandler.CardPlayedHandler{},
	model.GameEventType_CARD_REVEALED:      &eventhandler.CardRevealedHandler{},
	model.GameEventType_DAY_ADVANCED:       &eventhandler.DayAdvancedHandler{},
	model.GameEventType_LOOP_RESET:         &eventhandler.LoopResetHandler{},
	model.GameEventType_GAME_ENDED:         &eventhandler.GameOverHandler{},
	model.GameEventType_INCIDENT_TRIGGERED: &eventhandler.IncidentTriggeredHandler{},
	model.GameEventType_LOOP_WIN:           &eventhandler.LoopWinHandler{},
	model.GameEventType_LOOP_LOSS:          &eventhandler.LoopLossHandler{},
}

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

	gameConfig loader.GameConfig

	engineChan chan engineRequest
	stopChan   chan struct{}

	dispatchGameEvent chan *model.GameEvent

	currentPhase phase.Phase
	phaseTimer   *time.Timer
	gameStarted  bool

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
		currentPhase:         &phase.SetupPhase{}, // Start with the new SetupPhase
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
			case *aiActionCompleteRequest:
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

func (ge *GameEngine) processEvent(event *model.GameEvent) {
	// 1. Apply the state change using the appropriate handler
	if handler, ok := eventHandlers[event.Type]; ok {
		if err := handler.Handle(ge.GameState, event); err != nil {
			ge.logger.Error("Failed to handle event", zap.Error(err), zap.String("type", event.Type.String()))
		}
	} else {
		ge.logger.Warn("No handler registered for event type", zap.String("type", event.Type.String()))
	}

	// 2. Let the current phase react to the event
	nextPhase := ge.currentPhase.HandleEvent(ge, event)
	ge.transitionTo(nextPhase)

	// 3. Check for any new triggers that might have been activated
	// ge.checkForTriggers(event) // TODO: Re-implement trigger logic
}

func (ge *GameEngine) transitionTo(nextPhase phase.Phase) {
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
	ge.ApplyAndPublishEvent(model.GameEventType_GAME_ENDED, &model.GameOverEvent{Winner: winner})
	ge.logger.Info("Game over", zap.String("winner", winner.String()))
	ge.transitionTo(&phase.GameOverPhase{})
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

func (ge *GameEngine) GetGameConfig() loader.GameConfig {
	return ge.gameConfig
}

func (ge *GameEngine) ApplyAndPublishEvent(eventType model.GameEventType, payload proto.Message) {
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

	// First, process the event to apply state changes synchronously
	ge.processEvent(event)

	// Then, publish the event for external listeners
	ge.publishGameEvent(event)
}

func (ge *GameEngine) publishGameEvent(event *model.GameEvent) {
	select {
	case ge.dispatchGameEvent <- event:
		// Also record the event in the game state for player views
		ge.GameState.DayEvents = append(ge.GameState.DayEvents, event)
		ge.GameState.LoopEvents = append(ge.GameState.LoopEvents, event)
	default:
		ge.logger.Warn("Game event channel full, dropping event", zap.String("eventType", event.Type.String()))
	}
}

func (ge *GameEngine) AreAllPlayersReady() bool {
	//TODO: implement me
	return false
}

func (ge *GameEngine) ResolveMovement() {
	//TODO: implement me
}

func (ge *GameEngine) ResolveOtherCards() {
	//TODO: implement me
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
