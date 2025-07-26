package engine

import (
	"fmt"
	"time"
	promptbuilder "tragedylooper/internal/llm/prompt"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"tragedylooper/internal/game/data"
	"tragedylooper/internal/game/proto/model"
	"tragedylooper/internal/llm"
)

// GameEngine 管理单个游戏实例的状态和逻辑。
type GameEngine struct {
	GameState            *model.GameState
	requestChan          chan engineRequest
	gameEventChan        chan model.GameEvent
	gameControlChan      chan struct{}
	llmClient            llm.Client
	playerReady          map[string]bool
	mastermindPlayerID   string
	protagonistPlayerIDs []string
	logger               *zap.Logger
}

// engineRequest is an interface for all requests handled by the game engine loop.
type engineRequest interface{}

// getPlayerViewRequest is a request to get a filtered view of the game state for a player.
type getPlayerViewRequest struct {
	playerID     string
	responseChan chan model.PlayerView
}

// llmActionCompleteRequest is sent when an LLM player has decided on an action.
type llmActionCompleteRequest struct {
	playerID string
	action   model.PlayerAction
}

// NewGameEngine 创建一个新的游戏引擎实例。
func NewGameEngine(gameID string, logger *zap.Logger, script *model.Script, players map[string]*model.Player, llmClient llm.Client) *GameEngine {
	characters := make(map[string]*model.Character)
	for _, charConfig := range script.Characters {
		char := &model.Character{
			Id:              charConfig.CharacterId,
			Name:            charConfig.CharacterId, // Placeholder
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
			p.Hand = make([]model.Card, len(data.MastermindCards))
			copy(p.Hand, data.MastermindCards)
		} else {
			ge.protagonistPlayerIDs = append(ge.protagonistPlayerIDs, playerID)
			p.Hand = make([]model.Card, len(data.ProtagonistCards))
			copy(p.Hand, data.ProtagonistCards)
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

// GetPlayerView 为特定玩家生成游戏状态的过滤视图。
// It is thread-safe as it communicates with the main game loop via a channel.
func (ge *GameEngine) GetPlayerView(playerID string) model.PlayerView {
	responseChan := make(chan model.PlayerView)
	req := getPlayerViewRequest{
		playerID:     playerID,
		responseChan: responseChan,
	}

	// Send the request to the game loop and wait for the response.
	ge.requestChan <- req
	view := <-responseChan
	return view
}

func (ge *GameEngine) GetCharacter(id string) (*model.Character, bool) {
	char, ok := ge.GameState.Characters[id]
	return char, ok
}

func (ge *GameEngine) SetCharacterLocation(id string, location model.LocationType) {
	if char, ok := ge.GameState.Characters[id]; ok {
		char.CurrentLocation = location
		ge.logger.Info("Character moved", zap.String("characterID", id), zap.String("location", string(location)))
		ge.publishGameEvent(model.EventCharacterMoved, map[string]string{"character_id": id, "new_location": string(location)})
	}
}

func (ge *GameEngine) AdjustCharacterParanoia(id string, amount int) int {
	if char, ok := ge.GameState.Characters[id]; ok {
		char.Paranoia += amount
		ge.logger.Info("Character paranoia adjusted", zap.String("characterID", id), zap.Int("amount", amount), zap.Int("newParanoia", char.Paranoia))
		ge.publishGameEvent(model.EventParanoiaAdjusted, map[string]interface{}{"character_id": id, "amount": amount, "new_paranoia": char.Paranoia})
		return char.Paranoia
	}
	return 0
}

func (ge *GameEngine) AdjustCharacterGoodwill(id string, amount int) int {
	if char, ok := ge.GameState.Characters[id]; ok {
		char.Goodwill += amount
		ge.logger.Info("Character goodwill adjusted", zap.String("characterID", id), zap.Int("amount", amount), zap.Int("newGoodwill", char.Goodwill))
		ge.publishGameEvent(model.EventGoodwillAdjusted, map[string]interface{}{"character_id": id, "amount": amount, "new_goodwill": char.Goodwill})
		return char.Goodwill
	}
	return 0
}

func (ge *GameEngine) AdjustCharacterIntrigue(id string, amount int) int {
	if char, ok := ge.GameState.Characters[id]; ok {
		char.Intrigue += amount
		ge.logger.Info("Character intrigue adjusted", zap.String("characterID", id), zap.Int("amount", amount), zap.Int("newIntrigue", char.Intrigue))
		ge.publishGameEvent(model.EventIntrigueAdjusted, map[string]interface{}{"character_id": id, "amount": amount, "new_intrigue": char.Intrigue})
		return char.Intrigue
	}
	return 0
}

func (ge *GameEngine) PublishEvent(eventType model.EventType, payload interface{}) {
	ge.publishGameEvent(eventType, payload)
}

// --- Main Game Loop ---

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
			case model.PlayerAction:
				ge.handlePlayerAction(r)
			case getPlayerViewRequest:
				playerView := ge.generatePlayerView(r.playerID)
				r.responseChan <- playerView
			case llmActionCompleteRequest:
				// The LLM has finished. Process its action and mark it as ready.
				ge.handlePlayerAction(r.action)
				ge.playerReady[r.playerID] = true
			}

		case <-timer.C:
			// Advance the game state based on the current phase.
			switch ge.GameState.CurrentPhase {
			case model.PhaseMorning:
				ge.handleMorningPhase()
			case model.PhaseCardPlay:
				ge.handleCardPlayPhase()
			case model.PhaseCardReveal:
				ge.handleCardRevealPhase()
			case model.PhaseCardResolve:
				ge.handleCardResolvePhase()
			case model.PhaseAbilities:
				ge.handleAbilitiesPhase()
			case model.PhaseIncidents:
				ge.handleIncidentsPhase()
			case model.PhaseDayEnd:
				ge.handleDayEndPhase()
			case model.PhaseLoopEnd:
				ge.handleLoopEndPhase()
			case model.PhaseProtagonistGuess:
				ge.handleProtagonistGuessPhase()
			case model.PhaseGameOver:
				// Do nothing, wait for StopGameLoop
			}
		}
	}
}

// --- Player Action Handlers ---

func (ge *GameEngine) handlePlayerAction(action model.PlayerAction) {
	player, ok := ge.GameState.Players[action.PlayerID]
	if !ok {
		ge.logger.Warn("Action from unknown player", zap.String("playerID", action.PlayerID))
		return
	}

	ge.logger.Info("Handling player action", zap.String("player", player.Name), zap.String("actionType", string(action.Type)))

	switch action.Type {
	case model.ActionPlayCard:
		ge.handlePlayCardAction(player, action)
	case model.ActionUseAbility:
		ge.handleUseAbilityAction(player, action)
	case model.ActionReadyForNextPhase:
		ge.handleReadyForNextPhaseAction(player)
	case model.ActionMakeGuess:
		ge.handleMakeGuessAction(action)
	default:
		ge.logger.Warn("Unknown action type", zap.String("actionType", string(action.Type)))
	}
}

func (ge *GameEngine) handlePlayCardAction(player *model.Player, action model.PlayerAction) {
	var payload model.PlayCardPayload
	if err := mapstructure.Decode(action.Payload, &payload); err != nil {
		ge.logger.Error("Failed to decode PlayCardPayload", zap.Error(err))
		return
	}

	var playedCard model.Card
	cardFound := false
	for i, card := range player.Hand {
		if card.ID == payload.CardID {
			if card.OncePerLoop && card.UsedThisLoop {
				ge.logger.Warn("Attempted to play a card that was already used this loop", zap.String("cardID", card.ID))
				return // Card already used
			}
			playedCard = card
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...) // Remove card from hand
			cardFound = true
			break
		}
	}

	if !cardFound {
		ge.logger.Warn("Attempted to play a card not in hand", zap.String("cardID", payload.CardID), zap.String("playerID", player.ID))
		return
	}

	// Add target info to the card instance before storing it
	playedCard.TargetCharacterID = payload.TargetCharacterID
	playedCard.TargetLocation = payload.TargetLocation
	playedCard.UsedThisLoop = true // Mark as used

	ge.GameState.PlayedCardsThisDay[player.ID] = append(ge.GameState.PlayedCardsThisDay[player.ID], playedCard)
	ge.GameState.PlayedCardsThisLoop[player.ID] = append(ge.GameState.PlayedCardsThisLoop[player.ID], playedCard)
	ge.playerReady[player.ID] = true
}

func (ge *GameEngine) handleUseAbilityAction(player *model.Player, action model.PlayerAction) {
	var payload model.UseAbilityPayload
	if err := mapstructure.Decode(action.Payload, &payload); err != nil {
		ge.logger.Error("Failed to decode UseAbilityPayload", zap.Error(err))
		return
	}

	var ability *model.Ability
	abilityFound := false
	char, ok := ge.GameState.Characters[payload.CharacterID]
	if !ok {
		ge.logger.Warn("Character not found for ability use", zap.String("characterID", payload.CharacterID))
		return
	}

	for i := range char.Abilities {
		if char.Abilities[i].ID == payload.AbilityID {
			ability = &char.Abilities[i]
			abilityFound = true
			break
		}
	}

	if !abilityFound {
		ge.logger.Warn("Ability not found on character", zap.String("abilityID", payload.AbilityID), zap.String("characterID", payload.CharacterID))
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

func (ge *GameEngine) handleReadyForNextPhaseAction(player *model.Player) {
	ge.playerReady[player.ID] = true
}

func (ge *GameEngine) handleMakeGuessAction(action model.PlayerAction) {
	if ge.GameState.CurrentPhase != model.PhaseProtagonistGuess {
		ge.logger.Warn("MakeGuess action received outside of the guess phase")
		return
	}

	var payload model.MakeGuessPayload
	if err := mapstructure.Decode(action.Payload, &payload); err != nil {
		ge.logger.Error("Failed to decode MakeGuessPayload", zap.Error(err))
		return
	}

	correctGuesses := 0
	totalCharactersToGuess := 0
	for charID, guessedRole := range payload.GuessedRoles {
		char, exists := ge.GameState.Characters[charID]
		if !exists {
			continue // Ignore guesses for non-existent characters
		}
		// Only count characters that have a hidden role to be guessed
		if char.HiddenRole != "" {
			totalCharactersToGuess++
			if char.HiddenRole == guessedRole {
				correctGuesses++
			}
		}
	}

	if totalCharactersToGuess > 0 && correctGuesses == totalCharactersToGuess {
		ge.endGame(model.PlayerRoleProtagonist)
	} else {
		ge.endGame(model.PlayerRoleMastermind)
	}
}

// --- Game Phase Handlers ---

func (ge *GameEngine) handleMorningPhase() {
	ge.logger.Info("Morning Phase", zap.Int("loop", ge.GameState.CurrentLoop), zap.Int("day", ge.GameState.CurrentDay))
	ge.resetPlayerReadiness()
	ge.GameState.PlayedCardsThisDay = make(map[string][]model.Card) // Clear cards for the new day

	// Trigger DayStart abilities
	for _, char := range ge.GameState.Characters {
		for i, ability := range char.Abilities {
			if ability.TriggerType == model.AbilityTriggerDayStart && !ability.UsedThisLoop {
				payload := model.UseAbilityPayload{CharacterID: char.ID, AbilityID: ability.ID} // Assuming self-target for simplicity
				if err := ge.applyEffect(ability.Effect, &ability, payload); err != nil {
					ge.logger.Error("Error applying DayStart ability effect", zap.Error(err), zap.String("character", char.Name), zap.String("ability", ability.Name))
				}
				ge.GameState.Characters[char.ID].Abilities[i].UsedThisLoop = true // Mark as used
				ge.publishGameEvent(model.EventAbilityUsed, map[string]string{"character_id": char.ID, "ability_name": ability.Name})
			}
		}
	}

	ge.GameState.CurrentPhase = model.PhaseCardPlay
	ge.publishGameEvent(model.EventDayAdvanced, map[string]int{"day": ge.GameState.CurrentDay, "loop": ge.GameState.CurrentLoop})
}

func (ge *GameEngine) handleCardPlayPhase() {
	allPlayersReady := true
	for playerID, player := range ge.GameState.Players {
		if ge.playerReady[playerID] {
			continue
		}
		if player.IsLLM {
			go ge.triggerLLMPlayerAction(playerID)
		}
		allPlayersReady = false
	}

	if allPlayersReady {
		ge.logger.Info("All players ready for Card Reveal.")
		ge.GameState.CurrentPhase = model.PhaseCardReveal
	}
}

func (ge *GameEngine) handleCardRevealPhase() {
	ge.logger.Info("Card Reveal Phase")
	ge.publishGameEvent(model.EventCardPlayed, ge.GameState.PlayedCardsThisDay)
	ge.GameState.CurrentPhase = model.PhaseCardResolve
}

func (ge *GameEngine) handleCardResolvePhase() {
	ge.logger.Info("Card Resolve Phase")
	// Resolve cards in a specific order if necessary (e.g., by initiative). For now, iterate over players.
	for playerID, cards := range ge.GameState.PlayedCardsThisDay {
		for _, card := range cards {
			// We need to create a payload that fits the new UseAbilityPayload structure.
			// However, card effects are not directly tied to a character's ability in the same way.
			// This part of the logic might need a bigger refactor depending on how card effects are intended to work.
			// For now, we'll pass an empty payload and adjust the applyEffect function if necessary.
			// A better approach would be to have card effects not use UseAbilityPayload, but their own struct.
			payload := model.UseAbilityPayload{} // This is a temporary fix.
			if err := ge.applyEffect(card.Effect, nil, payload); err != nil {
				ge.logger.Error("Error applying card effect",
					zap.Error(err),
					zap.String("playerID", playerID),
					zap.String("cardName", card.Name),
				)
			}
		}
	}
	ge.GameState.CurrentPhase = model.PhaseAbilities
}

func (ge *GameEngine) handleAbilitiesPhase() {
	ge.logger.Info("Abilities Phase")
	// This phase can be complex, involving player choices.
	// For now, we assume a simple flow where the Mastermind might use an ability.
	mastermindPlayer := ge.GameState.Players[ge.mastermindPlayerID]
	if !ge.playerReady[mastermindPlayer.ID] {
		if mastermindPlayer.IsLLM {
			go ge.triggerLLMPlayerAction(mastermindPlayer.ID)
		}
		return // Wait for mastermind action/readiness
	}

	// After mastermind, protagonists might act. This requires more state and player interaction.
	// For now, we'll just move to the next phase.
	ge.resetPlayerReadiness()
	ge.GameState.CurrentPhase = model.PhaseIncidents
}

func (ge *GameEngine) handleIncidentsPhase() {
	ge.logger.Info("Incidents Phase")
	tragedyOccurred := false
	for _, tragedy := range ge.GameState.Script.Tragedies {
		// Check if the tragedy is active for the day and hasn't been prevented
		if tragedy.Day == ge.GameState.CurrentDay && ge.GameState.ActiveTragedies[tragedy.TragedyType] && !ge.GameState.PreventedTragedies[tragedy.TragedyType] {
			if ge.checkTragedyConditions(tragedy) {
				ge.logger.Info("Tragedy triggered!", zap.String("tragedy_type", string(tragedy.TragedyType)))
				ge.publishGameEvent(model.EventTragedyTriggered, map[string]string{"tragedy_type": string(tragedy.TragedyType)})
				tragedyOccurred = true
				break // Only one tragedy per day
			}
		}
	}

	if tragedyOccurred {
		ge.GameState.CurrentPhase = model.PhaseLoopEnd
	} else {
		ge.GameState.CurrentPhase = model.PhaseDayEnd
	}
}

func (ge *GameEngine) handleDayEndPhase() {
	ge.logger.Info("Day End Phase", zap.Int("day", ge.GameState.CurrentDay))
	ge.GameState.CurrentDay++
	if ge.GameState.CurrentDay > ge.GameState.Script.DaysPerLoop {
		ge.GameState.CurrentPhase = model.PhaseLoopEnd
	} else {
		ge.GameState.CurrentPhase = model.PhaseMorning
	}
}

func (ge *GameEngine) handleLoopEndPhase() {
	ge.logger.Info("Loop End Phase", zap.Int("loop", ge.GameState.CurrentLoop))

	// Check for Mastermind win condition (a tragedy occurred or final loop ended with un-prevented tragedies)
	mastermindWins := false
	for _, tragedy := range ge.GameState.Script.Tragedies {
		if ge.GameState.ActiveTragedies[tragedy.TragedyType] && !ge.GameState.PreventedTragedies[tragedy.TragedyType] {
			// This check is broad. A more precise check would be if a tragedy *actually occurred* this loop.
			// For now, we assume any un-prevented tragedy at loop end is a win condition.
			mastermindWins = true
			break
		}
	}

	// If it's the last loop, the outcome is final.
	if ge.GameState.CurrentLoop >= ge.GameState.Script.LoopCount {
		if mastermindWins {
			ge.endGame(model.PlayerRoleMastermind)
		} else {
			ge.endGame(model.PlayerRoleProtagonist)
		}
		return
	}

	// If a tragedy occurred mid-loop, Mastermind wins immediately.
	if mastermindWins { // Simplified check, should be based on an actual event.
		ge.endGame(model.PlayerRoleMastermind)
		return
	}

	// If no tragedy occurred and more loops are left, reset for the next loop.
	ge.resetLoop()
	ge.GameState.CurrentLoop++
	ge.GameState.CurrentDay = 1
	ge.GameState.CurrentPhase = model.PhaseMorning
	ge.publishGameEvent(model.EventLoopReset, map[string]int{"loop": ge.GameState.CurrentLoop})
}

func (ge *GameEngine) handleProtagonistGuessPhase() {
	ge.logger.Info("Protagonist Guess Phase")
	// This phase is triggered by a player action (ActionMakeGuess).
	// The logic is handled in `handleMakeGuessAction`.
	// After the guess, the game transitions to GameOver.
	ge.GameState.CurrentPhase = model.PhaseGameOver
}

// --- Core Logic & Helper Functions ---

func (ge *GameEngine) processEvent(event model.Event) {
	// This function applies the consequences of a resolved effect event to the game state.
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

	// First, see if the effect requires a choice from the player.
	choices, err := effect.ResolveChoices(ctx, ability)
	if err != nil {
		return fmt.Errorf("error resolving choices: %w", err)
	}

	// If choices are available and no specific target was provided in the payload, ask the player.
	if len(choices) > 1 && payload.CharacterID == "" { // Simplified check, might need more robust logic
		ge.publishGameEvent(model.EventChoiceRequired, choices)
		return nil // Stop processing and wait for a player action with the choice.
	}

	// If a choice was made or not required, execute the effect.
	events, err := effect.Execute(ctx, ability, payload)
	if err != nil {
		return fmt.Errorf("error executing effect: %w", err)
	}

	// Apply all resulting events to the game state.
	for _, event := range events {
		ge.processEvent(event)
	}

	return nil
}

// checkTragedyConditions checks if the conditions for a given tragedy are met.
func (ge *GameEngine) checkTragedyConditions(tragedy model.TragedyCondition) bool {
	for _, cond := range tragedy.Conditions {
		char, ok := ge.GameState.Characters[cond.CharacterID]
		if !ok || !char.IsAlive {
			return false // Character not found or not alive
		}
		if char.CurrentLocation != cond.Location {
			return false // Location mismatch
		}
		if char.Paranoia < cond.MinParanoia {
			return false // Paranoia too low
		}
		if cond.IsAlone {
			countAtLocation := 0
			for _, otherChar := range ge.GameState.Characters {
				if otherChar.CurrentLocation == cond.Location && otherChar.IsAlive {
					countAtLocation++
				}
			}
			if countAtLocation > 1 {
				return false // Not alone
			}
		}
	}
	return true // All conditions met
}

// resetLoop resets the game state for a new loop.
func (ge *GameEngine) resetLoop() {
	ge.logger.Info("Resetting for new loop...")
	// Reset characters to their initial script configuration
	for _, charConfig := range ge.GameState.Script.Characters {
		if char, ok := ge.GameState.Characters[charConfig.CharacterID]; ok {
			char.CurrentLocation = charConfig.InitialLocation
			char.Paranoia = 0
			char.Goodwill = 0
			char.Intrigue = 0
			char.IsAlive = true
			for i := range char.Abilities {
				char.Abilities[i].UsedThisLoop = false
			}
		}
	}
	// Reset card usage status for all players
	for _, player := range ge.GameState.Players {
		for i := range player.Hand {
			player.Hand[i].UsedThisLoop = false
		}
	}

	// Clear loop-specific state
	ge.GameState.PreventedTragedies = make(map[model.TragedyType]bool)
	ge.GameState.PlayedCardsThisDay = make(map[string][]model.Card)
	ge.GameState.PlayedCardsThisLoop = make(map[string][]model.Card)
	ge.GameState.DayEvents = []model.GameEvent{}
	ge.GameState.LoopEvents = []model.GameEvent{}

	ge.logger.Info("Loop reset complete.")
}

// endGame transitions the game to a finished state.
func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.logger.Info("Game Over!", zap.String("winner", string(winner)))
	ge.GameState.CurrentPhase = model.PhaseGameOver
	ge.publishGameEvent(model.EventGameOver, map[string]string{"winner": string(winner)})
	ge.StopGameLoop()
}

// resetPlayerReadiness resets the ready status for all players.
func (ge *GameEngine) resetPlayerReadiness() {
	for playerID := range ge.playerReady {
		ge.playerReady[playerID] = false
	}
}

// generatePlayerView creates a filtered view of the game state for a specific player.
// This method is NOT thread-safe and must only be called from within the runGameLoop goroutine.
func (ge *GameEngine) generatePlayerView(playerID string) model.PlayerView {
	player := ge.GameState.Players[playerID]
	if player == nil {
		return model.PlayerView{} // Or handle error
	}

	view := model.PlayerView{
		GameID:             ge.GameState.GameID,
		ScriptID:           ge.GameState.Script.ID,
		CurrentDay:         ge.GameState.CurrentDay,
		CurrentLoop:        ge.GameState.CurrentLoop,
		CurrentPhase:       ge.GameState.CurrentPhase,
		ActiveTragedies:    ge.GameState.ActiveTragedies,
		PreventedTragedies: ge.GameState.PreventedTragedies,
		PublicEvents:       ge.GameState.DayEvents,
	}

	// Filter characters based on player role
	view.Characters = make(map[string]*model.Character)
	for id, char := range ge.GameState.Characters {
		charCopy := *char // Create a copy to avoid modifying original data
		if player.Role == model.PlayerRoleProtagonist {
			charCopy.HiddenRole = "" // Hide role from protagonists
		}
		view.Characters[id] = &charCopy
	}

	// Filter player info
	view.Players = make(map[string]*model.Player)
	for id, p := range ge.GameState.Players {
		playerCopy := *p
		if id != playerID {
			playerCopy.Hand = nil // Hide other players' hands
		}
		view.Players[id] = &playerCopy
	}

	// Add player-specific info
	view.YourHand = player.Hand
	if player.Role == model.PlayerRoleProtagonist {
		view.YourDeductions = player.DeductionKnowledge
	}

	return view
}

// --- LLM Integration ---

// triggerLLMPlayerAction prompts an LLM player to make a decision.
func (ge *GameEngine) triggerLLMPlayerAction(playerID string) {
	player := ge.GameState.Players[playerID]
	if player == nil || !player.IsLLM {
		return
	}

	ge.logger.Info("Triggering LLM for player", zap.String("player", player.Name), zap.String("role", string(player.Role)))
	playerView := ge.GetPlayerView(playerID) // Get a safe, filtered view of the game state
	pBuilder := promptbuilder.NewPromptBuilder()
	var prompt string
	if player.Role == model.PlayerRoleMastermind {
		prompt = pBuilder.BuildMastermindPrompt(playerView, ge.GameState.Script, ge.GameState.Characters)
	} else {
		prompt = pBuilder.BuildProtagonistPrompt(playerView, player.DeductionKnowledge)
	}

	go func() {
		llmResponse, err := ge.llmClient.GenerateResponse(prompt, player.LLMSessionID)
		if err != nil {
			ge.logger.Error("LLM call failed", zap.String("player", player.Name), zap.Error(err))
			// Submit a default action to unblock the game
			ge.requestChan <- model.PlayerAction{PlayerID: playerID, GameID: ge.GameState.GameID, Type: model.ActionReadyForNextPhase}
			return
		}

		responseParser := llm.NewResponseParser()
		llmAction, err := responseParser.ParseLLMAction(llmResponse)
		if err != nil {
			ge.logger.Error("Failed to parse LLM response", zap.String("player", player.Name), zap.Error(err))
			// Submit a default action to unblock the game
			ge.requestChan <- model.PlayerAction{PlayerID: playerID, GameID: ge.GameState.GameID, Type: model.ActionReadyForNextPhase}
			return
		}

		// Here, a symbolic AI component could validate or refine the LLM's suggestion.
		// This is the core of the "Hybrid AI" approach. For now, we trust the LLM's action.

		// Send the validated action back to the main loop for processing.
		ge.requestChan <- llmActionCompleteRequest{
			playerID: playerID,
			action:   llmAction,
		}
	}()
}

func (ge *GameEngine) publishGameEvent(eventType model.EventType, payload interface{}) {
	event := model.GameEvent{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	select {
	case ge.gameEventChan <- event:
		// Also record the event in the game state for player views
		ge.GameState.DayEvents = append(ge.GameState.DayEvents, event)
		ge.GameState.LoopEvents = append(ge.GameState.LoopEvents, event)
	default:
		ge.logger.Warn("Game event channel full, dropping event", zap.String("eventType", string(eventType)))
	}
}
