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
	GameState            *model.GameState
	gameConfig           loader.GameConfigAccessor
	requestChan          chan engineRequest
	dispatchGameEvent    chan *model.GameEvent
	stopChan             chan struct{}
	llmClient            llm.Client
	playerReady          map[int32]bool
	mastermindPlayerID   int32
	protagonistPlayerIDs []int32
	characterNameToID    map[string]int32
	logger               *zap.Logger
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
func NewGameEngine(gameID string, logger *zap.Logger, players map[int32]*model.Player, llmClient llm.Client, gameConfig loader.GameConfigAccessor) *GameEngine {
	gs := &model.GameState{
		GameId:                  gameID,
		Characters:              make(map[int32]*model.Character),
		Players:                 players,
		CurrentDay:              1,
		CurrentLoop:             1,
		CurrentPhase:            model.GamePhase_SETUP,
		ActiveTragedies:         make(map[int32]bool),
		PreventedTragedies:      make(map[int32]bool),
		PlayedCardsThisDay:      make(map[int32]*model.Card),
		PlayedCardsThisLoop:     make(map[int32]bool),
		LastUpdateTime:          time.Now().Unix(),
		DayEvents:               make([]*model.GameEvent, 0),
		LoopEvents:              make([]*model.GameEvent, 0),
		CharacterParanoiaLimits: make(map[int32]int32),
		CharacterGoodwillLimits: make(map[int32]int32),
		CharacterIntrigueLimits: make(map[int32]int32),
	}

	for _, charConfig := range gameConfig.GetCharacters() {
		character := &model.Character{Config: charConfig}
		gs.Characters[charConfig.Id] = character
	}

	ge := &GameEngine{
		GameState:         gs,
		gameConfig:        gameConfig,
		requestChan:       make(chan engineRequest, 100),
		dispatchGameEvent: make(chan *model.GameEvent, 100),
		stopChan:          make(chan struct{}),
		llmClient:         llmClient,
		playerReady:       make(map[int32]bool),
		characterNameToID: make(map[string]int32),
		logger:            logger.With(zap.String("gameID", gameID)),
	}

	for id, char := range gs.Characters {
		ge.characterNameToID[char.Config.Name] = id
	}

	for playerID, p := range players {
		switch p.Role {
		case model.PlayerRole_MASTERMIND:
			ge.mastermindPlayerID = playerID
		case model.PlayerRole_PROTAGONIST:
			ge.protagonistPlayerIDs = append(ge.protagonistPlayerIDs, playerID)
		default:
			ge.logger.Warn("Unknown player role", zap.Int32("playerID", playerID))
		}
	}

	return ge
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
	req := getPlayerViewRequest{
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
				ge.playerReady[r.playerID] = true
			case getPlayerViewRequest:
				r.responseChan <- ge.createPlayerView(r.playerID)
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
	ge.applyAndPublishEvent(model.GameEventType_LOOP_OVER, &model.GameOverEvent{Winner: winner})
	ge.logger.Info("Game over", zap.String("winner", winner.String()))
}

func (ge *GameEngine) resetPlayerReadiness() {
	for playerID := range ge.GameState.Players {
		ge.playerReady[playerID] = false
	}
}

func (ge *GameEngine) resetLoop() {
	for _, char := range ge.GameState.Characters {
		char.Paranoia = 0
		char.Goodwill = 0
		char.Intrigue = 0
	}
	for _, p := range ge.GameState.Players {
		p.Hand = nil
	}
	ge.GameState.PlayedCardsThisLoop = make(map[int32]bool)
	ge.GameState.PreventedTragedies = make(map[int32]bool)
	ge.GameState.DayEvents = make([]*model.GameEvent, 0)
}

func (ge *GameEngine) createPlayerView(playerID int32) *model.PlayerView {
	player, ok := ge.GameState.Players[playerID]
	if !ok {
		return nil // Or return an empty view
	}

	view := &model.PlayerView{
		GameId:             ge.GameState.GameId,
		CurrentDay:         ge.GameState.CurrentDay,
		CurrentLoop:        ge.GameState.CurrentLoop,
		CurrentPhase:       ge.GameState.CurrentPhase,
		ActiveTragedies:    ge.GameState.ActiveTragedies,
		PreventedTragedies: ge.GameState.PreventedTragedies,
		YourHand:           player.Hand,
		YourDeductions:     player.DeductionKnowledge,
		PublicEvents:       ge.GameState.DayEvents, // A simplified version, might need filtering
		Characters:         make(map[int32]*model.PlayerViewCharacter),
		Players:            make(map[int32]*model.PlayerViewPlayer),
	}

	// Populate character views
	for id, char := range ge.GameState.Characters {
		view.Characters[id] = &model.PlayerViewCharacter{
			Id:              id,
			Name:            char.Config.Name,
			Traits:          char.Traits,
			CurrentLocation: char.CurrentLocation,
			Paranoia:        char.Paranoia,
			Goodwill:        char.Goodwill,
			Intrigue:        char.Intrigue,
			Abilities:       char.Abilities,
			IsAlive:         char.IsAlive,
			InPanicMode:     char.InPanicMode,
			Rules:           char.Config.Rules,
		}
	}

	// Populate player views
	for id, p := range ge.GameState.Players {
		view.Players[id] = &model.PlayerViewPlayer{
			Id:   id,
			Name: p.Name,
			Role: p.Role, // Mastermind will see all roles, Protagonists might see their own
		}
	}

	// If the player is not the mastermind, hide secret information

	return view
}
