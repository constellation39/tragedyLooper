package engine

import (
	"fmt"
	"time"
	promptbuilder "tragedylooper/internal/llm/prompt"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"tragedylooper/internal/game/model"
	"tragedylooper/internal/llm"
)

// --- GameMutator Implementation ---

// Statically assert that *GameEngine implements the model.GameMutator interface.
var _ model.GameMutator = (*GameEngine)(nil)

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

// getPlayerViewRequest is a request to get a filtered view of the game state for a player.
type getPlayerViewRequest struct {
	playerID     string
	responseChan chan model.PlayerView
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
		PublicEvents:       ge.GameState.DayEvents, // Or LoopEvents, depending on context
	}

	// Filter characters: Mastermind sees all, Protagonists don't see HiddenRole
	view.Characters = make(map[string]*model.Character)
	for id, char := range ge.GameState.Characters {
		charCopy := *char // Create a copy to avoid modifying original data
		if player.Role == model.PlayerRoleProtagonist {
			charCopy.HiddenRole = "" // Hide role
		}
		view.Characters[id] = &charCopy
	}

	// Filter players: Hide other players' hands
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

// --- 游戏阶段处理函数（示例） ---

func (ge *GameEngine) handleMorningPhase() {
	ge.logger.Info("Morning Phase", zap.Int("loop", ge.GameState.CurrentLoop), zap.Int("day", ge.GameState.CurrentDay))
	// 重置玩家准备状态
	for playerID := range ge.playerReady {
		ge.playerReady[playerID] = false
	}
	ge.GameState.PlayedCardsThisDay = make(map[string][]model.Card) // 清空当天打出的牌
	// 触发 DayStart 能力
	for _, char := range ge.GameState.Characters {
		for i, ability := range char.Abilities {
			if ability.TriggerType == model.AbilityTriggerDayStart && !ability.UsedThisLoop {
				// 假设能力效果直接应用，没有目标。如果能力需要目标，则需要更复杂的逻辑。
				payload := model.UseAbilityPayload{TargetCharacterID: char.ID}
				if err := ge.applyEffect(ability.Effect, &ability, payload); err != nil {
					ge.logger.Error("Error applying DayStart ability effect", zap.Error(err), zap.String("character", char.Name), zap.String("ability", ability.Name))
				}
				ge.GameState.Characters[char.ID].Abilities[i].UsedThisLoop = true // 标记为已使用
				ge.publishGameEvent(model.EventAbilityUsed, map[string]string{"character_id": char.ID, "ability_name": ability.Name})
			}
		}
	}
	ge.GameState.CurrentPhase = model.PhaseCardPlay
	ge.publishGameEvent(model.EventDayAdvanced, map[string]int{"day": ge.GameState.CurrentDay, "loop": ge.GameState.CurrentLoop})
}

func (ge *GameEngine) handleCardPlayPhase() {
	// 检查所有玩家（人类和 LLM）是否已打出牌或已准备好
	allPlayersReady := true
	for playerID, player := range ge.GameState.Players {
		if player.IsLLM {
			// 对于 LLM 玩家，触发其决策
			if !ge.playerReady[playerID] {
				go ge.triggerLLMPlayerAction(playerID)
				allPlayersReady = false // 等待 LLM 响应
			}
		} else {
			// 对于人类玩家，检查他们是否已打出牌或明确标记为准备就绪
			if !ge.playerReady[playerID] {
				allPlayersReady = false
			}
		}
	}
	if allPlayersReady {
		ge.logger.Info("All players ready for Card Reveal.")
		ge.GameState.CurrentPhase = model.PhaseCardReveal
	}
}

func (ge *GameEngine) handleCardRevealPhase() {
	ge.logger.Info("Card Reveal Phase")
	// 广播当天所有打出的牌
	ge.publishGameEvent(model.EventCardPlayed, ge.GameState.PlayedCardsThisDay)
	ge.GameState.CurrentPhase = model.PhaseCardResolve
}

func (ge *GameEngine) handleCardResolvePhase() {
	ge.logger.Info("Card Resolve Phase")
	// 按顺序解析所有打出的牌
	for playerID, cards := range ge.GameState.PlayedCardsThisDay {
		for _, card := range cards {
			// 从卡牌动作的载荷中获取目标信息
			// 注意：这假设卡牌在被打出时其目标信息被存储或可访问。
			// 在当前结构中，我们可能需要调整 PlayerAction 来包含这些信息，
			// 或者在 PlayedCardsThisDay 中存储更丰富的对象。
			// 为简单起见，我们假设目标信息在卡牌效果中可用或可以推断。
			payload := model.UseAbilityPayload{
				TargetCharacterID: card.TargetCharacterID,
				TargetLocation:    card.TargetLocation,
			}
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
	// 主谋能力（如果有）
	mastermindPlayer := ge.GameState.Players[ge.mastermindPlayerID]
	if mastermindPlayer.IsLLM {
		if !ge.playerReady[mastermindPlayer.ID] {
			go ge.triggerLLMPlayerAction(mastermindPlayer.ID) // LLM 决定是否使用能力
			return                                            // 等待 LLM
		}
	} else {
		if !ge.playerReady[mastermindPlayer.ID] {
			return
		}
	}
	// 善意能力（主角）
	// 这是一个复杂的阶段：领导者选择，主谋批准/拒绝（如果适用）
	// 为简单起见，我们假设一个基本流程。
	// 在实际游戏中，这将涉及更多的玩家动作和状态。

	ge.resetPlayerReadiness()
	ge.GameState.CurrentPhase = model.PhaseIncidents
}

func (ge *GameEngine) handleIncidentsPhase() {
	ge.logger.Info("Incidents Phase")
	// 检查悲剧是否发生
	tragedyOccurred := false
	for _, tragedy := range ge.GameState.Script.Tragedies {
		if tragedy.Day == ge.GameState.CurrentDay && ge.GameState.ActiveTragedies[tragedy.TragedyType] && !ge.GameState.PreventedTragedies[tragedy.TragedyType] {
			if ge.checkTragedyConditions(tragedy) {
				ge.logger.Info("Tragedy triggered!", zap.String("tragedy_type", string(tragedy.TragedyType)))
				ge.publishGameEvent(model.EventTragedyTriggered, map[string]string{"tragedy_type": string(tragedy.TragedyType)})
				tragedyOccurred = true
				break // 每天只能发生一个悲剧
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

	mastermindWins := false
	if ge.GameState.CurrentLoop >= ge.GameState.Script.LoopCount {
		// 所有循环用尽，如果任何活跃的悲剧未被阻止，则主谋获胜
		for _, tragedy := range ge.GameState.Script.Tragedies {
			if ge.GameState.ActiveTragedies[tragedy.TragedyType] && !ge.GameState.PreventedTragedies[tragedy.TragedyType] {
				mastermindWins = true
				break
			}
		}
		if mastermindWins {
			ge.endGame(model.PlayerRoleMastermind)
			return
		}
		ge.endGame(model.PlayerRoleProtagonist)
		return
	}

	for _, tragedy := range ge.GameState.Script.Tragedies {
		if ge.GameState.ActiveTragedies[tragedy.TragedyType] && !ge.GameState.PreventedTragedies[tragedy.TragedyType] {
			ge.endGame(model.PlayerRoleMastermind)
			return
		}
	}

	// 如果没有悲剧发生且循环未用尽，则重置以进行下一个循环
	ge.resetLoop()
	ge.GameState.CurrentLoop++
	ge.GameState.CurrentDay = 1
	ge.GameState.CurrentPhase = model.PhaseMorning
	ge.publishGameEvent(model.EventLoopReset, map[string]int{"loop": ge.GameState.CurrentLoop})
}

func (ge *GameEngine) handleProtagonistGuessPhase() {
	ge.logger.Info("Protagonist Guess Phase")
	// 这个阶段由主角决定进行最终猜测时触发。
	// 它需要一个特定的玩家动作 (ActionMakeGuess)。
	ge.GameState.CurrentPhase = model.PhaseGameOver // 猜测后过渡
}

// --- 核心游戏逻辑函数 ---

// checkTragedyConditions 检查给定悲剧的条件是否满足。
func (ge *GameEngine) checkTragedyConditions(tragedy model.TragedyCondition) bool {
	for _, cond := range tragedy.Conditions {
		char, ok := ge.GameState.Characters[cond.CharacterID]
		if !ok || !char.IsAlive {
			return false
		} // 角色未找到或未存活
		if char.CurrentLocation != cond.Location {
			return false
		} // 检查位置
		if char.Paranoia < cond.MinParanoia {
			return false
		}                 // 检查偏执
		if cond.IsAlone { // 检查是否独自在地点
			countAtLocation := 0
			for _, otherChar := range ge.GameState.Characters {
				if otherChar.CurrentLocation == cond.Location && otherChar.IsAlive {
					countAtLocation++
				}
			}
			if countAtLocation > 1 {
				return false
			}
		}
	}
	return true // 所有条件满足
}

// resetLoop 重置游戏状态以进行新循环。
func (ge *GameEngine) resetLoop() {
	ge.logger.Info("Resetting for new loop...")
	// 重置角色位置和状态到初始剧本配置
	for _, charConfig := range ge.GameState.Script.Characters {
		char := ge.GameState.Characters[charConfig.CharacterID]
		if char != nil {
			char.CurrentLocation = charConfig.InitialLocation
			char.Paranoia = 0   // 重置偏执
			char.Goodwill = 0   // 重置善意
			char.Intrigue = 0   // 重置阴谋
			char.IsAlive = true // 复活
			// 重置每循环一次能力的使用状态
			for i := range char.Abilities {
				char.Abilities[i].UsedThisLoop = false
			}
		}
	}
	// 重置每循环一次卡牌的使用状态
	for playerID := range ge.GameState.Players {
		player := ge.GameState.Players[playerID]
		for i := range player.Hand {
			player.Hand[i].UsedThisLoop = false
		}
	}
	// 清除新循环中被阻止的悲剧
	ge.GameState.PreventedTragedies = make(map[model.TragedyType]bool)
	ge.GameState.PlayedCardsThisDay = make(map[string][]model.Card)
	ge.GameState.PlayedCardsThisLoop = make(map[string][]model.Card) // 清除循环中所有打出的牌
	ge.GameState.DayEvents = []model.GameEvent{}                     // 清除日事件
	ge.GameState.LoopEvents = []model.GameEvent{}                    // 清除循环事件

	ge.logger.Info("Loop reset complete.")
}

// endGame 处理游戏结束状态。
func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.logger.Info("Game Over!", zap.String("winner", string(winner)))
	ge.GameState.CurrentPhase = model.PhaseGameOver
	ge.publishGameEvent(model.EventGameOver, map[string]string{"winner": string(winner)})
	ge.StopGameLoop() // 停止游戏循环
}

// resetPlayerReadiness 重置所有玩家的准备状态。
func (ge *GameEngine) resetPlayerReadiness() {
	for playerID := range ge.playerReady {
		ge.playerReady[playerID] = false
	}
}

// markCardUsed 标记卡牌在此循环中已使用。
/* func (ge *GameEngine) markCardUsed(cardID string) {
	for _, player := range ge.GameState.Players {
		for i := range player.Hand {
			if player.Hand[i].ID == cardID {
				player.Hand[i].UsedThisLoop = true
				return
			}
		}
	}
} */

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
			playedCard = card
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			cardFound = true
			break
		}
	}
	if !cardFound {
		return
	}
	if playedCard.OncePerLoop && playedCard.UsedThisLoop {
		return
	}

	// Add target info to the card instance before storing it
	playedCard.TargetCharacterID = payload.TargetCharacterID
	playedCard.TargetLocation = payload.TargetLocation

	ge.GameState.PlayedCardsThisDay[player.ID] = append(ge.GameState.PlayedCardsThisDay[player.ID], playedCard)
	ge.GameState.PlayedCardsThisLoop[player.ID] = append(ge.GameState.PlayedCardsThisLoop[player.ID], playedCard)
	ge.playerReady[player.ID] = true
}

func (ge *GameEngine) handleReadyForNextPhaseAction(player *model.Player) {
	ge.playerReady[player.ID] = true
}

func (ge *GameEngine) handleMakeGuessAction(action model.PlayerAction) {
	if ge.GameState.CurrentPhase != model.PhaseProtagonistGuess {
		return
	}
	payload, ok := action.Payload.(map[string]interface{})
	if !ok {
		return
	}
	guessedRolesMap, ok := payload["guessed_roles"].(map[string]interface{})
	if !ok {
		return
	}

	correctGuesses := 0
	totalCharacters := 0
	for charID, guessedRoleIfc := range guessedRolesMap {
		guessedRole, ok := guessedRoleIfc.(string)
		if !ok {
			continue
		}
		char, exists := ge.GameState.Characters[charID]
		if exists {
			totalCharacters++
			if char.HiddenRole == model.RoleType(guessedRole) {
				correctGuesses++
			}
		}
	}

	if correctGuesses == totalCharacters && totalCharacters > 0 {
		ge.endGame(model.PlayerRoleProtagonist)
	} else {
		ge.endGame(model.PlayerRoleMastermind)
	}
}

// --- LLM 集成逻辑（混合 AI） ---

// triggerLLMPlayerAction 提示 LLM 玩家做出决策。
func (ge *GameEngine) triggerLLMPlayerAction(playerID string) {
	player := ge.GameState.Players[playerID]
	if player == nil || !player.IsLLM {
		return
	}

	ge.logger.Info("Triggering LLM for player", zap.String("player", player.Name), zap.String("role", string(player.Role)))
	// We need to get the player view by sending a request to the main loop, not directly.
	playerView := ge.GetPlayerView(playerID)
	pBuilder := promptbuilder.NewPromptBuilder()
	prompt := ""
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
			ge.requestChan <- model.PlayerAction{
				PlayerID: playerID, GameID: ge.GameState.GameID, Type: model.ActionReadyForNextPhase, Payload: nil,
			}
			return
		}

		responseParser := llm.NewResponseParser()
		llmAction, err := responseParser.ParseLLMAction(llmResponse)
		if err != nil {
			ge.logger.Error("Failed to parse LLM response", zap.String("player", player.Name), zap.Error(err))
			// Submit a default action to unblock the game
			ge.requestChan <- model.PlayerAction{
				PlayerID: playerID, GameID: ge.GameState.GameID, Type: model.ActionReadyForNextPhase, Payload: nil,
			}
			return
		}

		// --- Hybrid AI Logic (Symbolic AI Component) ---
		// Here, a symbolic AI component would refine or validate the LLM's suggestion.
		// For example:
		// 1. LLM suggests playing Card X on Character Y.
		// 2. Symbolic AI (Go code) checks if Card X is in hand, if Character Y is a valid target,
		//    if the "once per loop" condition is met, etc.
		// 3. If the LLM's suggestion is invalid, the symbolic AI might:
		//    a. Correct it to a valid action.
		//    b. Ask the LLM to re-evaluate (if the API supports conversational turns).
		//    c. Fallback to a default, valid action (e.g., "pass turn").
		// 4. For strategic depth, especially for the Mastermind:
		//    The symbolic AI could run a small search (e.g., Monte Carlo Tree Search, simple decision tree)
		//    to find the optimal move given the current state and mastermind goals.
		//    The LLM would then be used to "narrate" or "justify" this optimal move,
		//    or to choose between strategically equivalent moves in a "human-like" way.
		//    For Protagonists, the symbolic AI could perform logical deduction based on observed events,
		//    updating `player.DeductionKnowledge` before prompting the LLM.

		// Instead of modifying state directly, send a request back to the main loop.
		ge.requestChan <- llmActionCompleteRequest{
			playerID: playerID,
			action: model.PlayerAction{
				PlayerID: playerID,
				GameID:   ge.GameState.GameID,
				Type:     llmAction.Type,
				Payload:  llmAction.Payload,
			},
		}
	}()
}

// engineRequest is an interface for all requests handled by the game engine loop.
type engineRequest interface{}

// llmActionCompleteRequest is sent when an LLM player has decided on an action.
type llmActionCompleteRequest struct {
	playerID string
	action   model.PlayerAction
}

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
			currentPhase := ge.GameState.CurrentPhase

			switch currentPhase {
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
			}
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

	if action.Type == model.ActionUseAbility {
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
