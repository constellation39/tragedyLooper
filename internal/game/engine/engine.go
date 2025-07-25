package engine

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"tragedylooper/internal/game/model"
	"tragedylooper/internal/llm"
)

// engineRequest is an interface for all requests handled by the game engine loop.
type engineRequest interface{}

// llmActionCompleteRequest is sent when an LLM player has decided on an action.
type llmActionCompleteRequest struct {
	playerID string
	action   model.PlayerAction
}

// GameEngine 管理单个游戏实例的状态和逻辑。
type GameEngine struct {
	GameState *model.GameState
	requestChan chan engineRequest
	gameEventChan chan model.GameEvent
	gameControlChan chan struct{}
	llmClient llm.Client
	playerReady map[string]bool
	mastermindPlayerID string
	protagonistPlayerIDs []string
	logger               *zap.Logger
}

// NewGameEngine 创建一个新的游戏引擎实例。
func NewGameEngine(gameID string, logger *zap.Logger, script model.Script, players map[string]*model.Player, llmClient llm.Client) *GameEngine {
	characters := make(map[string]*model.Character)
	for _, charConfig := range script.Characters {
		char := &model.Character{
			ID:              charConfig.CharacterID,
			Name:            charConfig.CharacterID, // Placeholder
			CurrentLocation: charConfig.InitialLocation,
			IsAlive:         true,
			HiddenRole:      charConfig.HiddenRole,
			Abilities:       []model.Ability{}, // Load from data
			Traits:          []string{},        // Load from data
		}
		characters[char.ID] = char
	}

	gs := &model.GameState{
		GameID:              gameID,
		Script:              script,
		Characters:          characters,
		Players:             players,
		CurrentDay:          1,
		CurrentLoop:         1,
		CurrentPhase:        model.PhaseMorning,
		ActiveTragedies:     make(map[model.TragedyType]bool),
		PreventedTragedies:  make(map[model.TragedyType]bool),
		PlayedCardsThisDay:  make(map[string][]model.Card),
		PlayedCardsThisLoop: make(map[string][]model.Card),
		LastUpdateTime:      time.Now(),
		DayEvents:           []model.GameEvent{},
		LoopEvents:          []model.GameEvent{},
	}

	for _, t := range script.Tragedies {
		gs.ActiveTragedies[t.TragedyType] = true
	}

	ge := &GameEngine{
		GameState:       gs,
		requestChan:     make(chan engineRequest, 100),
		gameEventChan:   make(chan model.GameEvent, 100),
		gameControlChan: make(chan struct{}),
		llmClient:       llmClient,
		playerReady:     make(map[string]bool),
		logger:          logger.With(zap.String("gameID", gameID)),
	}

	for playerID, p := range players {
		if p.Role == model.PlayerRoleMastermind {
			ge.mastermindPlayerID = playerID
		} else {
			ge.protagonistPlayerIDs = append(ge.protagonistPlayerIDs, playerID)
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

func (ge *GameEngine) SubmitPlayerAction(action model.PlayerAction) {
	select {
	case ge.requestChan <- action:
	default:
		ge.logger.Warn("Request channel full, dropping action", zap.String("playerID", action.PlayerID))
	}
}

func (ge *GameEngine) GetGameEvents() <-chan model.GameEvent {
	return ge.gameEventChan
}

// GetPlayerView is a temporary method to satisfy the server's dependency.
// In a true event-sourced architecture, this would be replaced by a proper
// read model or a state projection maintained by the client.
func (ge *GameEngine) GetPlayerView(playerID string) model.PlayerView {
	ge.logger.Warn("GetPlayerView is a deprecated method and should be refactored.")
	return model.PlayerView{}
}

func (ge *GameEngine) runGameLoop() {
	ge.logger.Info("Game loop started.")
	defer ge.logger.Info("Game loop stopped.")

	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ge.gameControlChan:
			return

		case req := <-ge.requestChan:
			switch r := req.(type) {
			case model.PlayerAction:
				ge.handlePlayerAction(r)
			case llmActionCompleteRequest:
				ge.handlePlayerAction(r.action)
				ge.playerReady[r.playerID] = true
			}

		case <-timer.C:
			// Phase handling logic would be here
		}
	}
}

// --- Core Logic ---

func (ge *GameEngine) processEvent(event model.Event) {
	switch e := event.(type) {
	case model.CharacterMovedEvent:
		if char, ok := ge.GameState.Characters[e.CharacterID]; ok {
			char.CurrentLocation = e.NewLocation
			ge.publishGameEvent(model.EventCharacterMoved, e)
		}
	case model.ParanoiaAdjustedEvent:
		if char, ok := ge.GameState.Characters[e.CharacterID]; ok {
			char.Paranoia += e.Amount
			ge.publishGameEvent(model.EventParanoiaAdjusted, e)
		}
	case model.GoodwillAdjustedEvent:
		if char, ok := ge.GameState.Characters[e.CharacterID]; ok {
			char.Goodwill += e.Amount
			ge.publishGameEvent(model.EventGoodwillAdjusted, e)
		}
	case model.IntrigueAdjustedEvent:
		if char, ok := ge.GameState.Characters[e.CharacterID]; ok {
			char.Intrigue += e.Amount
			ge.publishGameEvent(model.EventIntrigueAdjusted, e)
		}
	default:
		ge.logger.Warn("Unknown event type for processing", zap.Any("event", event))
	}
}

func (ge *GameEngine) applyEffect(effect model.Effect, ability *model.Ability, payload model.UseAbilityPayload) error {
	ctx := model.EffectContext{GameState: ge.GameState}

	choices, err := effect.ResolveChoices(ctx, ability)
	if err != nil {
		return fmt.Errorf("error resolving choices: %w", err)
	}

	if len(choices) > 1 && payload.TargetCharacterID == "" { // Simplified check
		ge.publishGameEvent(model.EventChoiceRequired, choices)
		return nil // Waiting for player choice
	}

	events, err := effect.Execute(ctx, ability, payload)
	if err != nil {
		return fmt.Errorf("error executing effect: %w", err)
	}

	for _, event := range events {
		ge.processEvent(event)
	}

	return nil
}

func (ge *GameEngine) handlePlayerAction(action model.PlayerAction) {
	player := ge.GameState.Players[action.PlayerID]
	if player == nil {
		ge.logger.Warn("Action from unknown player", zap.String("playerID", action.PlayerID))
		return
	}

	ge.logger.Info("Handling player action", zap.String("player", player.Name), zap.String("actionType", string(action.Type)))

	switch action.Type {
	case model.ActionUseAbility:
		ge.handleUseAbilityAction(player, action)
	// ... other actions
	}
}

func (ge *GameEngine) handleUseAbilityAction(player *model.Player, action model.PlayerAction) {
	var payload model.UseAbilityPayload
	if err := mapstructure.Decode(action.Payload, &payload); err != nil {
		ge.logger.Error("Failed to decode UseAbilityPayload", zap.Error(err))
		return
	}

	var ability *model.Ability
	abilityFound := false
	for _, char := range ge.GameState.Characters {
		for i := range char.Abilities {
			if char.Abilities[i].Name == payload.AbilityName {
				ability = &char.Abilities[i]
				abilityFound = true
				break
			}
		}
		if abilityFound {
			break
		}
	}

	if !abilityFound {
		ge.logger.Warn("Ability not found", zap.String("abilityName", payload.AbilityName))
		return
	}

	if err := ge.applyEffect(ability.Effect, ability, payload); err != nil {
		ge.logger.Error("Failed to apply effect for ability", zap.String("abilityName", ability.Name), zap.Error(err))
		return
	}

	if ability.OncePerLoop {
		ability.UsedThisLoop = true
	}
}

func (ge *GameEngine) publishGameEvent(eventType model.EventType, payload interface{}) {
	event := model.GameEvent{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	select {
	case ge.gameEventChan <- event:
		ge.GameState.DayEvents = append(ge.GameState.DayEvents, event)
		ge.GameState.LoopEvents = append(ge.GameState.LoopEvents, event)
	default:
		ge.logger.Warn("Game event channel full", zap.String("eventType", string(eventType)))
	}
}
