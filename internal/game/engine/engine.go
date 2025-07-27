package engine

import (
	"slices"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	"tragedylooper/internal/game/data"
	model "tragedylooper/internal/game/proto/v1"
	"tragedylooper/internal/llm"
)

// GameEngine manages the state and logic of a single game instance.
type GameEngine struct {
	GameState            *model.GameState
	requestChan          chan engineRequest
	gameEventChan        chan *model.GameEvent
	gameControlChan      chan struct{}
	llmClient            llm.Client
	playerReady          map[int32]bool
	mastermindPlayerID   int32
	protagonistPlayerIDs []int32
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
	action   *model.PlayerAction
}

// NewGameEngine creates a new game engine instance.
func NewGameEngine(gameID int32, logger *zap.Logger, script *model.Script, players map[int32]*model.Player, llmClient llm.Client) *GameEngine {
	characters := make(map[int32]*model.Character)
	for _, charConfig := range script.Characters {
		char := &model.Character{
			Id:              charConfig.Id,
			Name:            charConfig.Name, // Placeholder
			CurrentLocation: charConfig.InitialLocation,
			IsAlive:         true,
			HiddenRole:      charConfig.HiddenRole,
			Abilities:       make([]*model.Ability, 0), // Load from data
			Traits:          []string{},                // Load from data
		}
		characters[char.Id] = char
	}

	gs := &model.GameState{
		GameId:              gameID,
		Script:              script,
		Characters:          characters,
		Players:             players,
		CurrentDay:          1,
		CurrentLoop:         1,
		CurrentPhase:        model.GamePhase_GAME_PHASE_MORNING,
		ActiveTragedies:     make(map[int32]bool),
		PreventedTragedies:  make(map[int32]bool),
		PlayedCardsThisDay:  make(map[int32]*model.CardList),
		PlayedCardsThisLoop: make(map[int32]*model.CardList),
		LastUpdateTime:      timestamppb.Now(),
		DayEvents:           make([]*model.GameEvent, 0),
		LoopEvents:          make([]*model.GameEvent, 0),
	}

	for _, t := range script.Tragedies {
		gs.ActiveTragedies[int32(t.TragedyType)] = true
	}

	ge := &GameEngine{
		GameState:       gs,
		requestChan:     make(chan engineRequest, 100),
		gameEventChan:   make(chan *model.GameEvent, 100),
		gameControlChan: make(chan struct{}),
		llmClient:       llmClient,
		playerReady:     make(map[int32]bool),
		logger:          logger.With(zap.Int32("gameID", gameID)),
	}

	for playerID, p := range players {
		if p.Role == model.PlayerRole_PLAYER_ROLE_MASTERMIND {
			ge.mastermindPlayerID = playerID
			p.Hand = slices.Clone(data.MastermindCards)
		} else {
			ge.protagonistPlayerIDs = append(ge.protagonistPlayerIDs, playerID)
			p.Hand = slices.Clone(data.ProtagonistCards)
		}
	}

	return ge
}

func (ge *GameEngine) StartGameLoop() {
	go ge.runGameLoop()
}

func (ge *GameEngine) StopGameLoop() {
	close(ge.gameControlChan)
}

func (ge *GameEngine) SubmitPlayerAction(action *model.PlayerAction) {
	select {
	case ge.requestChan <- action:
	default:
		ge.logger.Warn("Request channel full, dropping action", zap.Int32("playerID", action.PlayerId))
	}
}

func (ge *GameEngine) GetGameEvents() <-chan *model.GameEvent {
	return ge.gameEventChan
}

// GetPlayerView generates a filtered view of the game state for a specific player.
// It is thread-safe as it communicates with the main game loop via a channel.
func (ge *GameEngine) GetPlayerView(playerID int32) *model.PlayerView {
	responseChan := make(chan *model.PlayerView)
	req := getPlayerViewRequest{
		playerID:     playerID,
		responseChan: responseChan,
	}

	// Send the request to the game loop and wait for the response.
	ge.requestChan <- req
	view := <-responseChan
	return view
}

func (ge *GameEngine) runGameLoop() {
	ge.logger.Info("Game loop started.")
	defer ge.logger.Info("Game loop stopped.")

	// This timer drives the phase transitions.
	timer := time.NewTicker(100 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ge.gameControlChan:
			return

		case req := <-ge.requestChan:
			switch r := req.(type) {
			case *model.PlayerAction:
				ge.handlePlayerAction(r)
			case *getPlayerViewRequest:
				playerView := ge.generatePlayerView(r.playerID)
				r.responseChan <- playerView
			case *llmActionCompleteRequest:
				// The LLM has finished. Process its action and mark it as ready.
				ge.handlePlayerAction(r.action)
				ge.playerReady[r.playerID] = true
			}

		case <-timer.C:
			// Advance the game state based on the current phase.
			switch ge.GameState.CurrentPhase {
			case model.GamePhase_GAME_PHASE_MORNING:
				ge.handleMorningPhase()
			case model.GamePhase_GAME_PHASE_CARD_PLAY:
				ge.handleCardPlayPhase()
			case model.GamePhase_GAME_PHASE_CARD_REVEAL:
				ge.handleCardRevealPhase()
			case model.GamePhase_GAME_PHASE_CARD_RESOLVE:
				ge.handleCardResolvePhase()
			case model.GamePhase_GAME_PHASE_ABILITIES:
				ge.handleAbilitiesPhase()
			case model.GamePhase_GAME_PHASE_INCIDENTS:
				ge.handleIncidentsPhase()
			case model.GamePhase_GAME_PHASE_DAY_END:
				ge.handleDayEndPhase()
			case model.GamePhase_GAME_PHASE_LOOP_END:
				ge.handleLoopEndPhase()
			case model.GamePhase_GAME_PHASE_PROTAGONIST_GUESS:
				ge.handleProtagonistGuessPhase()
			case model.GamePhase_GAME_PHASE_GAME_OVER:
				// Do nothing, wait for StopGameLoop
			}
		}
	}
}
