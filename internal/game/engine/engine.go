package engine

import (
	"context"
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// LocationGrid defines the 2x2 map layout.
var LocationGrid = map[model.LocationType]struct{ X, Y int }{
	model.LocationType_SHRINE:   {0, 0},
	model.LocationType_SCHOOL:   {1, 0},
	model.LocationType_HOSPITAL: {0, 1},
	model.LocationType_CITY:     {1, 1},
}

// GameEngine manages the state and logic of a single game instance.
type engineAction interface{}

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
	em              *eventManager

	// engineChan is the central channel for all incoming requests (player actions, AI actions, etc.).
	// It ensures that all modifications to the game state are processed sequentially in the main game loop,
	// preventing race conditions.
	engineChan chan engineAction
	stopChan   chan struct{}

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
		engineChan:           make(chan engineAction, 100),
		stopChan:             make(chan struct{}),
		playerReady:          make(map[int32]bool),
		mastermindPlayerID:   0,
		protagonistPlayerIDs: nil,
	}
	ge.pm = newPhaseManager(ge)
	ge.em = newEventManager(ge)

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
	return ge.em.eventsChannel()
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
	defer ge.em.close()

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
func (ge *GameEngine) handleEngineRequest(req engineAction) {
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

func (ge *GameEngine) ApplyAndPublishEvent(eventType model.GameEventType, payload proto.Message) {
	ge.em.createAndProcess(eventType, payload)
}

func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.logger.Info("Game over", zap.String("winner", winner.String()))
	// This event will be processed by the event manager, leading to a state update and a phase transition.
	ge.em.createAndProcess(model.GameEventType_GAME_ENDED, &model.GameOverEvent{Winner: winner})
}

func (ge *GameEngine) ResetPlayerReadiness() {
	for playerID := range ge.GameState.Players {
		ge.playerReady[playerID] = false
	}
}

func (ge *GameEngine) GetCharacterByID(charID int32) *model.Character {
	char, ok := ge.GameState.Characters[charID]
	if !ok {
		return nil
	}
	return char
}

func (ge *GameEngine) TriggerIncidents() {
	// TODO: Implement incident triggering logic
}

func (ge *GameEngine) MoveCharacter(char *model.Character, dx, dy int) {
	ge.moveCharacter(char, dx, dy)
}

func (ge *GameEngine) getPlayerByID(playerID int32) *model.Player {
	player, ok := ge.GameState.Players[playerID]
	if !ok {
		return nil
	}
	return player
}

func (ge *GameEngine) moveCharacter(char *model.Character, dx, dy int) {
	startPos, ok := LocationGrid[char.CurrentLocation]
	if !ok {
		ge.logger.Warn("character in unknown location", zap.String("char", char.Config.Name))
		return
	}

	// Calculate the new position, wrapping around the 2x2 grid.
	newX := (startPos.X + dx) % 2
	newY := (startPos.Y + dy) % 2

	var newLoc model.LocationType
	for loc, pos := range LocationGrid {
		if pos.X == newX && pos.Y == newY {
			newLoc = loc
			break
		}
	}

	if newLoc != model.LocationType_LOCATION_TYPE_UNSPECIFIED && newLoc != char.CurrentLocation {
		// Check for movement restrictions
		for _, rule := range char.Config.Rules {
			if smr, ok := rule.Effect.(*model.CharacterRule_SpecialMovementRule); ok {
				for _, restricted := range smr.SpecialMovementRule.RestrictedLocations {
					if restricted == newLoc {
						ge.logger.Info("character movement restricted", zap.String("char", char.Config.Name), zap.String("location", newLoc.String()))
						return // Movement forbidden
					}
				}
			}
		}

		char.CurrentLocation = newLoc
		ge.ApplyAndPublishEvent(model.GameEventType_CHARACTER_MOVED, &model.CharacterMovedEvent{
			CharacterId: char.Config.Id,
			NewLocation: newLoc,
		})
		ge.logger.Info("character moved", zap.String("char", char.Config.Name), zap.String("to", newLoc.String()))
	}
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

func (ge *GameEngine) ResolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, player *model.Player, payload *model.UseAbilityPayload, ability *model.Ability) ([]int32, error) {
	return []int32{}, nil
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
