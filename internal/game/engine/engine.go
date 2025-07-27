package engine

import (
	"time"
	"tragedylooper/internal/game/loader"

	"go.uber.org/zap"
	model "tragedylooper/internal/game/proto/v1"
	"tragedylooper/internal/llm"
)

// GameEngine manages the state and logic of a single game instance.
type GameEngine struct {
	GameState            *model.GameState
	gameData             loader.GameConfigAccessor
	requestChan          chan engineRequest
	gameEventChan        chan *model.GameEvent
	gameControlChan      chan struct{}
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
func NewGameEngine(gameID string, logger *zap.Logger, players map[int32]*model.Player, llmClient llm.Client, gameData loader.GameConfigAccessor) *GameEngine {
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

	for _, charConfig := range gameData.GetCharacters() {
		character := &model.Character{}

		gs.Characters[charConfig.Id] = character
	}

	ge := &GameEngine{
		GameState:         gs,
		gameData:          gameData,
		requestChan:       make(chan engineRequest, 100),
		gameEventChan:     make(chan *model.GameEvent, 100),
		gameControlChan:   make(chan struct{}),
		llmClient:         llmClient,
		playerReady:       make(map[int32]bool),
		characterNameToID: make(map[string]int32),
		logger:            logger.With(zap.String("gameID", gameID)),
	}

	for _, char := range gs.Characters {
		ge.characterNameToID[char.Name] = char.Id
	}

	for playerID, p := range players {
		if p.Role == model.PlayerRole_MASTERMIND {
			ge.mastermindPlayerID = playerID
			//p.Hand = slices.Clone(data.MastermindCards)
		} else {
			ge.protagonistPlayerIDs = append(ge.protagonistPlayerIDs, playerID)
			//p.Hand = slices.Clone(data.ProtagonistCards)
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
			case *llmActionCompleteRequest:
				// The LLM has finished. Process its action and mark it as ready.
				ge.handlePlayerAction(r.playerID, r.action)
				ge.playerReady[r.playerID] = true
			}

		case <-timer.C:
			// Advance the game state based on the current phase.
			switch ge.GameState.CurrentPhase {
			case model.GamePhase_SETUP:
				ge.handleMorningPhase()
			case model.GamePhase_MASTERMIND_SETUP:
			case model.GamePhase_CARD_PLAY:
				ge.handleCardPlayPhase()
			case model.GamePhase_CARD_REVEAL:
				ge.handleCardRevealPhase()
			case model.GamePhase_CARD_RESOLVE:
				ge.handleCardResolvePhase()
			case model.GamePhase_ABILITIES:
				ge.handleAbilitiesPhase()
			case model.GamePhase_INCIDENTS:
				ge.handleIncidentsPhase()
			case model.GamePhase_DAY_END:
				ge.handleDayEndPhase()
			case model.GamePhase_LOOP_END:
				ge.handleLoopEndPhase()
			case model.GamePhase_GAME_OVER:
				ge.handleProtagonistGuessPhase()
			case model.GamePhase_FIRST_GUESS:
			case model.GamePhase_PROTAGONIST_GUESS:
			case model.GamePhase_GAME_PHASE_UNSPECIFIED:
			}
		}
	}
}

func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.GameState.CurrentPhase = model.GamePhase_GAME_OVER
	ge.publishGameEvent(model.GameEventType_LOOP_OVER, &model.GameOverEvent{Winner: winner})
	ge.logger.Info("Game over", zap.String("winner", winner.String()))
}

func (ge *GameEngine) resetPlayerReadiness() {
	for playerID := range ge.GameState.Players {
		ge.playerReady[playerID] = false
	}
}

func (ge *GameEngine) resetLoop() {
	// Reset character states, used cards, etc.
	for _, char := range ge.GameState.Characters {
		// Reset paranoia, goodwill, intrigue, location based on script
		char.Paranoia = 0
		char.Goodwill = 0
		char.Intrigue = 0
	}
	for _, p := range ge.GameState.Players {
		p.Hand = nil // Or reset to initial cards
	}
	ge.GameState.PlayedCardsThisLoop = make(map[int32]bool)
	ge.GameState.PreventedTragedies = make(map[int32]bool)
	ge.GameState.DayEvents = make([]*model.GameEvent, 0)
}

func (ge *GameEngine) SetCharacterLocation(characterID int32, location model.LocationType) {
	if char, ok := ge.GameState.Characters[characterID]; ok {
		char.CurrentLocation = location
		ge.publishGameEvent(model.GameEventType_CHARACTER_MOVED, &model.CharacterMovedEvent{CharacterId: characterID, NewLocation: location})
	}
}

func (ge *GameEngine) AdjustCharacterParanoia(characterID int32, amount int32) {
	if char, ok := ge.GameState.Characters[characterID]; ok {
		char.Paranoia += amount
		ge.publishGameEvent(model.GameEventType_PARANOIA_ADJUSTED, &model.ParanoiaAdjustedEvent{CharacterId: characterID, NewParanoia: char.Paranoia, Amount: amount})
	}
}

func (ge *GameEngine) AdjustCharacterGoodwill(characterID int32, amount int32) {
	if char, ok := ge.GameState.Characters[characterID]; ok {
		char.Goodwill += amount
		ge.publishGameEvent(model.GameEventType_GOODWILL_ADJUSTED, &model.GoodwillAdjustedEvent{CharacterId: characterID, NewGoodwill: char.Goodwill, Amount: amount})
	}
}

func (ge *GameEngine) AdjustCharacterIntrigue(characterID int32, amount int32) {
	if char, ok := ge.GameState.Characters[characterID]; ok {
		char.Intrigue += amount
		ge.publishGameEvent(model.GameEventType_INTRIGUE_ADJUSTED, &model.IntrigueAdjustedEvent{CharacterId: characterID, NewIntrigue: char.Intrigue, Amount: amount})
	}
}
