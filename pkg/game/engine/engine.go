package engine

import (
	"fmt"
	"log"
	"sync"
	"time"

	"tragedylooper/pkg/game/model"
	"tragedylooper/pkg/llm"
)

// GameEngine 管理单个游戏实例的状态和逻辑。
type GameEngine struct {
	GameState *model.GameState
	// 用于传入玩家动作和内部游戏事件的通道
	playerActionChan chan model.PlayerAction
	gameEventChan    chan model.GameEvent
	// 用于发出游戏结束或循环结束信号的通道
	gameControlChan chan struct{}
	// LLM 客户端用于 AI 玩家
	llmClient llm.LLMClient
	// 保护 GameState 访问的互斥锁（尽管事件循环最小化了直接争用）
	mu sync.Mutex
	// 跟踪玩家准备进入下一阶段的映射
	playerReady map[string]bool
	// 主谋玩家 ID
	mastermindPlayerID string
	// 主角玩家 ID 列表
	protagonistPlayerIDs []string
}

// NewGameEngine 创建一个新的游戏引擎实例。
func NewGameEngine(gameID string, script model.Script, players map[string]*model.Player, llmClient llm.LLMClient) *GameEngine {
	// 根据剧本配置初始化角色
	characters := make(map[string]*model.Character)
	for _, charConfig := range script.Characters {
		// 在实际应用中，你将从主列表加载完整的角色定义
		// 暂时创建一个基本角色
		char := &model.Character{
			ID:              charConfig.CharacterID,
			Name:            charConfig.CharacterID, // 占位符名称
			CurrentLocation: charConfig.InitialLocation,
			IsAlive:         true,
			HiddenRole:      charConfig.HiddenRole,
			Abilities:       []model.Ability{}, // 从主角色数据加载能力
			Traits:          []string{},        // 从主角色数据加载特性
		}
		characters[char.ID] = char
	}

	// 初始化游戏状态
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

	// 根据剧本设置活跃悲剧
	for _, t := range script.Tragedies {
		gs.ActiveTragedies[t.TragedyType] = true
	}

	ge := &GameEngine{
		GameState:        gs,
		playerActionChan: make(chan model.PlayerAction, 100), // 带缓冲通道
		gameEventChan:    make(chan model.GameEvent, 100),
		gameControlChan:  make(chan struct{}),
		llmClient:        llmClient,
		playerReady:      make(map[string]bool),
	}

	// 识别主谋和主角
	for playerID, p := range players {
		if p.Role == model.PlayerRoleMastermind {
			ge.mastermindPlayerID = playerID
		} else {
			ge.protagonistPlayerIDs = append(ge.protagonistPlayerIDs, playerID)
		}
	}

	return ge
}

// StartGameLoop 在协程中启动主游戏循环。
func (ge *GameEngine) StartGameLoop() {
	go ge.runGameLoop()
}

// StopGameLoop 停止主游戏循环。
func (ge *GameEngine) StopGameLoop() {
	close(ge.gameControlChan)
}

// SubmitPlayerAction 允许外部组件（例如 WebSocket 处理程序）提交玩家动作。
func (ge *GameEngine) SubmitPlayerAction(action model.PlayerAction) {
	select {
	case ge.playerActionChan <- action:
		// 动作提交成功
	default:
		log.Printf("Game %s: Player action channel full, dropping action from %s", ge.GameState.GameID, action.PlayerID)
	}
}

// GetGameEvents 返回一个通道，用于接收游戏事件以进行广播。
func (ge *GameEngine) GetGameEvents() <-chan model.GameEvent {
	return ge.gameEventChan
}

// GetPlayerView 为特定玩家生成游戏状态的过滤视图。
func (ge *GameEngine) GetPlayerView(playerID string) model.PlayerView {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	player := ge.GameState.Players[playerID]
	if player == nil {
		return model.PlayerView{} // 或处理错误
	}

	view := model.PlayerView{
		GameID:             ge.GameState.GameID,
		ScriptID:           ge.GameState.Script.ID,
		CurrentDay:         ge.GameState.CurrentDay,
		CurrentLoop:        ge.GameState.CurrentLoop,
		CurrentPhase:       ge.GameState.CurrentPhase,
		ActiveTragedies:    ge.GameState.ActiveTragedies,
		PreventedTragedies: ge.GameState.PreventedTragedies,
		PublicEvents:       ge.GameState.DayEvents, // 或 LoopEvents，取决于上下文
	}

	// 过滤角色：主谋看到所有，主角看不到 HiddenRole
	view.Characters = make(map[string]*model.Character)
	for id, char := range ge.GameState.Characters {
		charCopy := *char // 创建一个副本以避免修改原始数据
		if player.Role == model.PlayerRoleProtagonist {
			charCopy.HiddenRole = "" // 隐藏角色
		}
		view.Characters[id] = &charCopy
	}

	// 过滤玩家：隐藏其他玩家的手牌
	view.Players = make(map[string]*model.Player)
	for id, p := range ge.GameState.Players {
		playerCopy := *p
		if id != playerID {
			playerCopy.Hand = nil // 隐藏其他玩家的手牌
		}
		view.Players[id] = &playerCopy
	}

	// 添加玩家特定信息
	view.YourHand = player.Hand
	if player.Role == model.PlayerRoleProtagonist {
		view.YourDeductions = player.DeductionKnowledge
	}

	return view
}

// runGameLoop 是单个游戏实例的主事件循环。
func (ge *GameEngine) runGameLoop() {
	log.Printf("Game %s: Game loop started.", ge.GameState.GameID)
	defer log.Printf("Game %s: Game loop stopped.", ge.GameState.GameID)

	for {
		select {
		case <-ge.gameControlChan:
			return // 收到停止信号
		case action := <-ge.playerActionChan:
			ge.handlePlayerAction(action)
		default:
			// 处理游戏阶段
			ge.mu.Lock() // 锁定以修改状态
			currentPhase := ge.GameState.CurrentPhase
			ge.mu.Unlock()

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
				// 游戏结束，等待停止信号
				time.Sleep(time.Second)
			}
			time.Sleep(100 * time.Millisecond) // 防止忙等待
		}
	}
}

// --- 游戏阶段处理函数（示例） ---

func (ge *GameEngine) handleMorningPhase() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	log.Printf("Game %s, Loop %d, Day %d: Morning Phase", ge.GameState.GameID, ge.GameState.CurrentLoop, ge.GameState.CurrentDay)
	// 重置玩家准备状态
	for playerID := range ge.playerReady {
		ge.playerReady[playerID] = false
	}
	ge.GameState.PlayedCardsThisDay = make(map[string][]model.Card) // 清空当天打出的牌
	// 触发 DayStart 能力
	for _, char := range ge.GameState.Characters {
		for _, ability := range char.Abilities {
			if ability.TriggerType == model.AbilityTriggerDayStart && !ability.UsedThisLoop {
				ge.applyAbilityEffect(ability.Effect, char.ID) // 应用效果
				ability.UsedThisLoop = true                    // 如果是每循环一次，则标记为已使用
				ge.publishGameEvent(model.EventAbilityUsed, map[string]string{"character_id": char.ID, "ability_name": ability.Name})
			}
		}
	}
	ge.GameState.CurrentPhase = model.PhaseCardPlay
	ge.publishGameEvent(model.EventDayAdvanced, map[string]int{"day": ge.GameState.CurrentDay, "loop": ge.GameState.CurrentLoop})
}

func (ge *GameEngine) handleCardPlayPhase() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
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
		log.Printf("Game %s: All players ready for Card Reveal.", ge.GameState.GameID)
		ge.GameState.CurrentPhase = model.PhaseCardReveal
	} else {
		// 仍在等待玩家。保持阶段为 CardPlay。
		// 考虑在此处添加超时机制。
	}
}

func (ge *GameEngine) handleCardRevealPhase() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	log.Printf("Game %s: Card Reveal Phase", ge.GameState.GameID)
	// 广播当天所有打出的牌
	ge.publishGameEvent(model.EventCardPlayed, ge.GameState.PlayedCardsThisDay)
	ge.GameState.CurrentPhase = model.PhaseCardResolve
}

func (ge *GameEngine) handleCardResolvePhase() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	log.Printf("Game %s: Card Resolve Phase", ge.GameState.GameID)
	// 1. 首先解析移动卡
	for _, playerPlayedCards := range ge.GameState.PlayedCardsThisDay {
		for _, card := range playerPlayedCards {
			if card.CardType == model.CardTypeMovement {
				ge.applyCardEffect(card.Effect, card.TargetCharacterID, card.TargetLocation)
				if card.OncePerLoop {
					ge.markCardUsed(card.ID)
				}
			}
		}
	}
	// 2. 解析其他卡牌类型（偏执、善意、阴谋、特殊）
	for _, playerPlayedCards := range ge.GameState.PlayedCardsThisDay {
		for _, card := range playerPlayedCards {
			if card.CardType != model.CardTypeMovement {
				ge.applyCardEffect(card.Effect, card.TargetCharacterID, card.TargetLocation)
				if card.OncePerLoop {
					ge.markCardUsed(card.ID)
				}
			}
		}
	}
	ge.GameState.CurrentPhase = model.PhaseAbilities
}

func (ge *GameEngine) handleAbilitiesPhase() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	log.Printf("Game %s: Abilities Phase", ge.GameState.GameID)
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
	for _, playerID := range ge.protagonistPlayerIDs {
		player := ge.GameState.Players[playerID]
		if player.IsLLM {
			// 触发 LLM 主角提出善意能力
			// 这将是 LLM 当天决策的一部分。
		}
	}
	ge.resetPlayerReadiness()
	ge.GameState.CurrentPhase = model.PhaseIncidents
}

func (ge *GameEngine) handleIncidentsPhase() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	log.Printf("Game %s: Incidents Phase", ge.GameState.GameID)
	// 检查悲剧是否发生
	tragedyOccurred := false
	for _, tragedy := range ge.GameState.Script.Tragedies {
		if tragedy.Day == ge.GameState.CurrentDay && ge.GameState.ActiveTragedies[tragedy.TragedyType] && !ge.GameState.PreventedTragedies[tragedy.TragedyType] {
			if ge.checkTragedyConditions(tragedy) {
				log.Printf("Game %s: Tragedy %s triggered!", ge.GameState.GameID, tragedy.TragedyType)
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
	ge.mu.Lock()
	defer ge.mu.Unlock()
	log.Printf("Game %s: Day %d End Phase", ge.GameState.GameID, ge.GameState.CurrentDay)
	ge.GameState.CurrentDay++
	if ge.GameState.CurrentDay > ge.GameState.Script.DaysPerLoop {
		ge.GameState.CurrentPhase = model.PhaseLoopEnd
	} else {
		ge.GameState.CurrentPhase = model.PhaseMorning
	}
}

func (ge *GameEngine) handleLoopEndPhase() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	log.Printf("Game %s: Loop %d End Phase", ge.GameState.GameID, ge.GameState.CurrentLoop)

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
		} else {
			ge.endGame(model.PlayerRoleProtagonist)
			return
		}
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
	ge.mu.Lock()
	defer ge.mu.Unlock()
	log.Printf("Game %s: Protagonist Guess Phase", ge.GameState.GameID)
	// 这个阶段由主角决定进行最终猜测时触发。
	// 它需要一个特定的玩家动作 (ActionMakeGuess)。
	ge.GameState.CurrentPhase = model.PhaseGameOver // 猜测后过渡
}

// --- 核心游戏逻辑函数 ---

// applyCardEffect 将卡牌效果应用于游戏状态。
func (ge *GameEngine) applyCardEffect(effect model.AbilityEffect, targetCharacterID string, targetLocation model.LocationType) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	switch effect.Type {
	case model.EffectTypeMoveCharacter:
		char, ok := ge.GameState.Characters[targetCharacterID]
		if !ok {
			return fmt.Errorf("character %s not found for move effect", targetCharacterID)
		}
		if targetLocation == "" {
			return fmt.Errorf("target location not specified for move effect")
		}
		char.CurrentLocation = targetLocation
		ge.publishGameEvent(model.EventCharacterMoved, map[string]string{"character_id": targetCharacterID, "new_location": string(targetLocation)})
		log.Printf("Character %s moved to %s", targetCharacterID, targetLocation)
	case model.EffectTypeAdjustParanoia:
		char, ok := ge.GameState.Characters[targetCharacterID]
		if !ok {
			return fmt.Errorf("character %s not found for paranoia effect", targetCharacterID)
		}
		amount, ok := effect.Params["amount"].(float64) // JSON Unmarshals numbers to float64
		if !ok {
			return fmt.Errorf("invalid amount parameter for paranoia effect")
		}
		char.Paranoia += int(amount)
		ge.publishGameEvent(model.EventParanoiaAdjusted, map[string]interface{}{"character_id": targetCharacterID, "amount": int(amount), "new_paranoia": char.Paranoia})
		log.Printf("Character %s paranoia adjusted by %d to %d", targetCharacterID, int(amount), char.Paranoia)
	case model.EffectTypeAdjustGoodwill:
		char, ok := ge.GameState.Characters[targetCharacterID]
		if !ok {
			return fmt.Errorf("character %s not found for goodwill effect", targetCharacterID)
		}
		amount, ok := effect.Params["amount"].(float64)
		if !ok {
			return fmt.Errorf("invalid amount parameter for goodwill effect")
		}
		char.Goodwill += int(amount)
		ge.publishGameEvent(model.EventGoodwillAdjusted, map[string]interface{}{"character_id": targetCharacterID, "amount": int(amount), "new_goodwill": char.Goodwill})
		log.Printf("Character %s goodwill adjusted by %d to %d", targetCharacterID, int(amount), char.Goodwill)
	case model.EffectTypeAdjustIntrigue:
		char, ok := ge.GameState.Characters[targetCharacterID]
		if !ok {
			return fmt.Errorf("character %s not found for intrigue effect", targetCharacterID)
		}
		amount, ok := effect.Params["amount"].(float64)
		if !ok {
			return fmt.Errorf("invalid amount parameter for intrigue effect")
		}
		char.Intrigue += int(amount)
		ge.publishGameEvent(model.EventIntrigueAdjusted, map[string]interface{}{"character_id": targetCharacterID, "amount": int(amount), "new_intrigue": char.Intrigue})
		log.Printf("Character %s intrigue adjusted by %d to %d", targetCharacterID, int(amount), char.Intrigue)
	default:
		return fmt.Errorf("unsupported effect type: %s", effect.Type)
	}
	return nil
}

// applyAbilityEffect 将能力效果应用于游戏状态。
func (ge *GameEngine) applyAbilityEffect(effect model.AbilityEffect, targetCharacterID string) error {
	return ge.applyCardEffect(effect, targetCharacterID, "") // 传递空位置（如果不适用）
}

// checkTragedyConditions 检查给定悲剧的条件是否满足。
func (ge *GameEngine) checkTragedyConditions(tragedy model.TragedyCondition) bool {
	ge.mu.Lock()
	defer ge.mu.Unlock()
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
		} // 检查偏执
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
	log.Printf("Game %s: Resetting for new loop...", ge.GameState.GameID)
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

	log.Printf("Game %s: Loop reset complete.", ge.GameState.GameID)
}

// endGame 处理游戏结束状态。
func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	log.Printf("Game %s: Game Over! Winner: %s", ge.GameState.GameID, winner)
	ge.GameState.CurrentPhase = model.PhaseGameOver
	ge.publishGameEvent(model.EventGameOver, map[string]string{"winner": string(winner)})
	ge.StopGameLoop() // 停止游戏循环
}

// publishGameEvent 向游戏事件通道发送事件以进行广播。
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
		log.Printf("Game %s: Game event channel full, dropping event %s", ge.GameState.GameID, eventType)
	}
}

// resetPlayerReadiness 重置所有玩家的准备状态。
func (ge *GameEngine) resetPlayerReadiness() {
	for playerID := range ge.playerReady {
		ge.playerReady[playerID] = false
	}
}

// markCardUsed 标记卡牌在此循环中已使用。
func (ge *GameEngine) markCardUsed(cardID string) {
	for _, player := range ge.GameState.Players {
		for i := range player.Hand {
			if player.Hand[i].ID == cardID {
				player.Hand[i].UsedThisLoop = true
				return
			}
		}
	}
}

// --- 玩家动作处理 ---

func (ge *GameEngine) handlePlayerAction(action model.PlayerAction) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	player := ge.GameState.Players[action.PlayerID]
	if player == nil {
		log.Printf("Game %s: Action from unknown player %s", ge.GameState.GameID, action.PlayerID)
		return
	}

	log.Printf("Game %s: Player %s submitted action %s in phase %s", ge.GameState.GameID, player.Name, action.Type, ge.GameState.CurrentPhase)

	switch action.Type {
	case model.ActionPlayCard:
		if ge.GameState.CurrentPhase != model.PhaseCardPlay {
			return
		}
		payload, ok := action.Payload.(map[string]interface{})
		if !ok {
			return
		}
		cardID, ok := payload["card_id"].(string)
		if !ok {
			return
		}

		var playedCard model.Card
		cardFound := false
		for i, card := range player.Hand {
			if card.ID == cardID {
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

		ge.GameState.PlayedCardsThisDay[player.ID] = append(ge.GameState.PlayedCardsThisDay[player.ID], playedCard)
		ge.GameState.PlayedCardsThisLoop[player.ID] = append(ge.GameState.PlayedCardsThisLoop[player.ID], playedCard)
		ge.playerReady[player.ID] = true

	case model.ActionUseAbility:
		if ge.GameState.CurrentPhase != model.PhaseAbilities {
			return
		}
		payload, ok := action.Payload.(map[string]interface{})
		if !ok {
			return
		}
		abilityName, ok := payload["ability_name"].(string)
		if !ok {
			return
		}
		targetCharID, _ := payload["target_character_id"].(string)

		var usedAbility model.Ability
		abilityFound := false
		for _, char := range ge.GameState.Characters {
			if string(char.HiddenRole) == string(player.Role) || player.Role == model.PlayerRoleMastermind {
				for i, ab := range char.Abilities {
					if ab.Name == abilityName {
						usedAbility = ab
						if ab.OncePerLoop {
							char.Abilities[i].UsedThisLoop = true
						}
						abilityFound = true
						break
					}
				}
			}
			if abilityFound {
				break
			}
		}
		if !abilityFound {
			return
		}

		ge.applyAbilityEffect(usedAbility.Effect, targetCharID)
		ge.playerReady[player.ID] = true

	case model.ActionReadyForNextPhase:
		ge.playerReady[player.ID] = true

	case model.ActionMakeGuess:
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

	default:
		log.Printf("Game %s: Unknown action type: %s", ge.GameState.GameID, action.Type)
	}
}

// --- LLM 集成逻辑（混合 AI） ---

// triggerLLMPlayerAction 提示 LLM 玩家做出决策。
func (ge *GameEngine) triggerLLMPlayerAction(playerID string) {
	player := ge.GameState.Players[playerID]
	if player == nil || !player.IsLLM {
		return
	}

	log.Printf("Game %s: Triggering LLM for player %s (%s)", ge.GameState.GameID, player.Name, player.Role)
	playerView := ge.GetPlayerView(playerID)
	promptBuilder := llm.NewPromptBuilder()
	prompt := ""
	if player.Role == model.PlayerRoleMastermind {
		prompt = promptBuilder.BuildMastermindPrompt(playerView, ge.GameState.Script, ge.GameState.Characters)
	} else {
		prompt = promptBuilder.BuildProtagonistPrompt(playerView, player.DeductionKnowledge)
	}

	go func() {
		llmResponse, err := ge.llmClient.GenerateResponse(prompt, player.LLMSessionID)
		if err != nil {
			log.Printf("Game %s: LLM call for player %s failed: %v", ge.GameState.GameID, player.Name, err)
			ge.SubmitPlayerAction(model.PlayerAction{
				PlayerID: playerID, GameID: ge.GameState.GameID, Type: model.ActionReadyForNextPhase, Payload: nil,
			})
			return
		}

		responseParser := llm.NewResponseParser()
		llmAction, err := responseParser.ParseLLMAction(llmResponse)
		if err != nil {
			log.Printf("Game %s: Failed to parse LLM response for player %s: %v", ge.GameState.GameID, player.Name, err)
			ge.SubmitPlayerAction(model.PlayerAction{
				PlayerID: playerID, GameID: ge.GameState.GameID, Type: model.ActionReadyForNextPhase, Payload: nil,
			})
			return
		}

		// --- 混合 AI 逻辑 (符号 AI 组件) ---
		// 在这里，符号 AI 组件将细化或验证 LLM 的建议。
		// 例如：
		// 1. LLM 建议对角色 Y 打出卡牌 X。
		// 2. 符号 AI (Go 代码) 检查卡牌 X 是否在手牌中，角色 Y 是否是有效目标，
		//    是否满足卡牌的“每循环一次”条件等。
		// 3. 如果 LLM 的建议无效，符号 AI 可能会：
		//    a. 将其更正为有效动作。
		//    b. 要求 LLM 重新评估（如果 API 支持对话轮次）。
		//    c. 回退到默认的有效动作（例如，“跳过回合”）。
		// 4. 对于战略深度，特别是主谋：
		//    符号 AI 可以运行一个小型搜索（例如，蒙特卡洛树搜索，简单决策树）
		//    以找到在给定当前状态和主谋目标下的最佳动作。
		//    然后 LLM 将用于“叙述”或“证明”这个最佳动作，
		//    或者以“类人”的方式在战略上等效的动作之间进行选择。
		//    对于主角，符号 AI 可以根据观察到的事件执行逻辑推理，
		//    在提示 LLM 之前更新 `player.DeductionKnowledge`。

		// 此示例中，我们假设 LLM 的动作在解析后直接提交。
		ge.SubmitPlayerAction(model.PlayerAction{
			PlayerID: playerID, GameID: ge.GameState.GameID, Type: llmAction.Type, Payload: llmAction.Payload,
		})
		ge.mu.Lock()
		ge.playerReady[playerID] = true
		ge.mu.Unlock()
	}()
}
