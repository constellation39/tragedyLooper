package engine

import (
	"context"
	"fmt"
	"github.com/constellation39/tragedyLooper/internal/game/engine/ai"
	"github.com/constellation39/tragedyLooper/internal/game/engine/character"
	"github.com/constellation39/tragedyLooper/internal/game/engine/condition"
	"github.com/constellation39/tragedyLooper/internal/game/engine/effecthandler"
	"github.com/constellation39/tragedyLooper/internal/game/engine/eventhandler"
	"github.com/constellation39/tragedyLooper/internal/game/engine/phasehandler"
	"github.com/constellation39/tragedyLooper/internal/game/engine/target"
	"github.com/constellation39/tragedyLooper/internal/game/loader"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// engineAction 是一个空接口，用于标记所有可以发送到游戏引擎主循环的请求类型。
type engineAction interface{}

// getPlayerViewRequest is a request to get a player's filtered view of the game state.
type getPlayerViewRequest struct {
	playerID     int32
	responseChan chan *model.PlayerView
}

// actionCompleteRequest signifies that an AI or player action has been completed and is ready for processing by the game engine.
type actionCompleteRequest struct {
	playerID int32
	action   *model.PlayerActionPayload
}

// getCurrentPhaseRequest is a request to safely get the current game phase.
type getCurrentPhaseRequest struct {
	responseChan chan model.GamePhase
}

// GameEngine manages the state and logic of a single game instance.
type GameEngine struct {
	GameState *model.GameState
	logger    *zap.Logger

	actionGenerator ai.ActionGenerator
	gameConfig      loader.GameConfig
	pm              *phasehandler.Manager
	em              *eventhandler.Manager

	engineChan chan engineAction
	stopChan   chan struct{}

	playerReady map[int32]bool

	mastermindPlayerID   int32
	protagonistPlayerIDs []int32
}

// NewGameEngine creates a new game engine instance.
func NewGameEngine(logger *zap.Logger, players []*model.Player, actionGenerator ai.ActionGenerator, gameConfig loader.GameConfig) (*GameEngine, error) {
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
	ge.pm = phasehandler.NewManager(ge)
	ge.em = eventhandler.NewManager(ge)

	ge.initializeGameStateFromScript(players)
	ge.dealInitialCards()

	return ge, nil
}

func (ge *GameEngine) initializePlayers(players []*model.Player) map[int32]*model.Player {
	playerMap := make(map[int32]*model.Player, len(players))
	ge.protagonistPlayerIDs = make([]int32, 0, len(players)-1)

	for _, player := range players {
		switch player.Role {
		case model.PlayerRole_PLAYER_ROLE_MASTERMIND:
			ge.mastermindPlayerID = player.Id
		case model.PlayerRole_PLAYER_ROLE_PROTAGONIST:
			ge.protagonistPlayerIDs = append(ge.protagonistPlayerIDs, player.Id)
		default:
			ge.logger.Warn("Unknown player role", zap.Int32("playerID", player.Id))
		}

		playerMap[player.Id] = player
	}
	return playerMap
}

func (ge *GameEngine) initializeGameStateFromScript(players []*model.Player) {
	playerMap := ge.initializePlayers(players)

	script := ge.gameConfig.GetScript()
	characterConfigs := ge.gameConfig.GetCharacters()
	incidentConfigs := ge.gameConfig.GetIncidents()

	// 将脚本事件合并到主事件配置列表中
	for _, incident := range script.Incidents {
		incidentConfigs[incident.Id] = incident
	}

	ge.GameState = &model.GameState{
		CurrentLoop:         1,
		CurrentDay:          1,
		Players:             playerMap,
		Characters:          make(map[int32]*model.Character),
		PlayedCardsThisLoop: make(map[int32]bool),
		PlayedCardsThisDay:  make(map[int32]*model.CardList),
		TriggeredIncidents:  make(map[string]bool),
	}

	for _, charInScript := range script.Characters {
		charConfig, ok := characterConfigs[charInScript.CharacterId]
		if !ok {
			ge.logger.Warn("Character in script not found in character config", zap.Int32("charID", charInScript.CharacterId))
			continue
		}

		ge.GameState.Characters[charInScript.CharacterId] = &model.Character{
			Config:          charConfig,
			HiddenRole:      charInScript.HiddenRole,
			CurrentLocation: charInScript.InitialLocation,
		}
	}
}

func (ge *GameEngine) dealInitialCards() {
	cardConfigs := ge.gameConfig.GetCards()

	for _, player := range ge.GameState.Players {
		player.Hand = &model.CardList{Cards: make([]*model.Card, 0, len(cardConfigs))}
		for _, cardConfig := range cardConfigs {
			if cardConfig.OwnerRole == player.Role {
				player.Hand.Cards = append(player.Hand.Cards, &model.Card{Config: cardConfig})
			}
		}
	}
}

// Start 启动游戏主循环。
func (ge *GameEngine) Start() {
	go ge.runGameLoop()
}

// Stop 停止游戏主循环。
func (ge *GameEngine) Stop() {
	close(ge.stopChan)
}

// SubmitPlayerAction 提交玩家操作到游戏引擎。
func (ge *GameEngine) SubmitPlayerAction(playerID int32, action *model.PlayerActionPayload) {
	if action == nil {
		ge.logger.Warn("Received nil action from player")
		return
	}
	select {
	case ge.engineChan <- &actionCompleteRequest{playerID: playerID, action: action}:
	default:
		ge.logger.Warn("Request channel full, dropping action", zap.Int32("playerID", playerID))
	}
}

// GetGameEvents 返回游戏事件的只读通道。
func (ge *GameEngine) GetGameEvents() <-chan *model.GameEvent {
	return ge.em.EventsChannel()
}

// GetPlayerView 获取指定玩家的游戏状态视图。
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

// GetCurrentPhase 安全地从引擎获取当前游戏阶段。
func (ge *GameEngine) GetCurrentPhase() model.GamePhase {
	responseChan := make(chan model.GamePhase)
	req := &getCurrentPhaseRequest{
		responseChan: responseChan,
	}
	ge.engineChan <- req
	return <-responseChan
}

// runGameLoop 是游戏引擎的核心。它是一个单线程循环，按顺序处理所有游戏事件
// 和状态更改，从而在没有复杂锁定的情况下确保线程安全。
func (ge *GameEngine) runGameLoop() {
	ge.logger.Info("Game loop started.")
	defer ge.logger.Info("Game loop stopped.")

	// 阶段管理器已启动，它将启动第一个阶段转换。
	ge.pm.Start()
	ge.ResetPlayerReadiness()
	defer ge.em.Close()

	for {
		select {
		case <-ge.stopChan:
			return
		case req := <-ge.engineChan:
			ge.handleEngineRequest(req)
		case <-ge.pm.Timer():
			if ge.pm.HandleTimeout() {
				ge.ResetPlayerReadiness()
			}
		}
	}
}

// handleEngineRequest 处理来自引擎通道的传入请求。
func (ge *GameEngine) handleEngineRequest(req engineAction) {
	switch r := req.(type) {
	case *actionCompleteRequest:
		player, ok := ge.GameState.Players[r.playerID]
		if !ok {
			ge.logger.Warn("Action from unknown player", zap.Int32("playerID", r.playerID))
			return
		}
		ge.SetPlayerReady(r.playerID)
		if ge.pm.HandleAction(player, r.action) {
			ge.ResetPlayerReadiness()
		}
	case *getPlayerViewRequest:
		r.responseChan <- ge.GeneratePlayerView(r.playerID)
	case *getCurrentPhaseRequest:
		r.responseChan <- ge.pm.CurrentPhase().Type()
	default:
		ge.logger.Warn("Unhandled request type in engine channel")
	}
}

func (ge *GameEngine) TriggerEvent(eventType model.GameEventType, payload *model.EventPayload) {
	event := &model.GameEvent{
		Type:      eventType,
		Timestamp: timestamppb.Now(),
		Payload:   payload,
	}

	// Step 1: Apply the event to the game state through the appropriate handler.
	if err := ge.em.ApplyEvent(event); err != nil {
		ge.logger.Error("failed to apply event", zap.String("event", event.Type.String()), zap.Error(err))
		return
	}

	// Step 2: 让当前阶段对事件做出反应。
	// 这是一个重要的钩子，允许一个阶段根据发生的事件来改变游戏流程（例如，转换到不同的阶段）。
	// 注意：大多数阶段使用默认的空实现，因此这个调用通常不执行任何操作。
	// 但它为需要响应特定事件的阶段提供了必要的扩展点。
	if ge.pm.HandleEvent(event) {
		ge.ResetPlayerReadiness()
	}

	// Step 3: Record the event in the game state for player review.
	gs := ge.GetGameState()
	gs.DayEvents = append(gs.DayEvents, event)
	gs.LoopEvents = append(gs.LoopEvents, event)

	// Step 4: Publish the event to external listeners.
	ge.em.Dispatch(event)
}

// ResetPlayerReadiness 重置所有玩家的准备状态。
func (ge *GameEngine) ResetPlayerReadiness() {
	for playerID := range ge.GameState.Players {
		ge.playerReady[playerID] = false
	}
}

// GetCharacterByID 根据角色ID获取角色对象。
func (ge *GameEngine) GetCharacterByID(charID int32) *model.Character {
	char, ok := ge.GameState.Characters[charID]
	if !ok {
		return nil
	}
	return char
}

func (ge *GameEngine) MoveCharacter(char *model.Character, dx, dy int) {
	character.MoveCharacter(ge.logger, ge, ge.GameState, char, dx, dy)
}

func (ge *GameEngine) CheckCondition(cond *model.Condition) (bool, error) {
	return condition.Check(ge.GameState, cond)
}

func (ge *GameEngine) ResolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, ctx *effecthandler.EffectContext) ([]int32, error) {
	return target.ResolveSelectorToCharacters(gs, sel, ctx)
}

// getPlayerByID 根据玩家ID获取玩家对象。
func (ge *GameEngine) getPlayerByID(playerID int32) *model.Player {
	player, ok := ge.GameState.Players[playerID]
	if !ok {
		return nil
	}
	return player
}

// GetGameState implement phases.GameEngine, return current game state.
func (ge *GameEngine) GetGameState() *model.GameState {
	return ge.GameState
}

func (ge *GameEngine) GetGameRepo() loader.GameConfig {
	return ge.gameConfig
}

func (ge *GameEngine) AreAllPlayersReady() bool {
	for playerID := range ge.GameState.Players {
		if !ge.playerReady[playerID] {
			return false
		}
	}
	return true
}

// Logger 返回游戏引擎的日志记录器。
func (ge *GameEngine) Logger() *zap.Logger {
	return ge.logger
}

// SetPlayerReady 设置指定玩家的准备状态为 true。
func (ge *GameEngine) SetPlayerReady(playerID int32) {
	ge.playerReady[playerID] = true
}

// GetProtagonistPlayers 返回主角玩家。
func (ge *GameEngine) GetProtagonistPlayers() []*model.Player {
	players := make([]*model.Player, 0, len(ge.protagonistPlayerIDs))
	for _, id := range ge.protagonistPlayerIDs {
		players = append(players, ge.getPlayerByID(id))
	}
	return players
}

// GetMastermindPlayer 返回主谋玩家。
func (ge *GameEngine) GetMastermindPlayer() *model.Player {
	return ge.getPlayerByID(ge.mastermindPlayerID)
}

// ApplyEffect 查找效果的适当处理程序，解析选项，然后应用效果。
func (ge *GameEngine) ApplyEffect(effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error {
	handler, err := effecthandler.GetEffectHandler(effect)
	if err != nil {
		return err
	}

	ctx := &effecthandler.EffectContext{
		Ability: ability,
		Payload: payload,
		Choice:  choice,
	}

	// 1. 解析选项
	choices, err := handler.ResolveChoices(ge, effect, ctx)
	if err != nil {
		return fmt.Errorf("error resolving choices: %w", err)
	}

	if len(choices) > 0 && choice == nil {
		choiceEvent := &model.ChoiceRequiredEvent{Choices: choices}
		ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_CHOICE_REQUIRED, &model.EventPayload{
			Payload: &model.EventPayload_ChoiceRequired{ChoiceRequired: choiceEvent},
		})
		return nil // 停止处理，直到做出选择
	}

	// 2. 应用效果
	err = handler.Apply(ge, effect, ctx)
	if err != nil {
		return fmt.Errorf("error applying effect: %w", err)
	}

	return nil
}

// RequestAIAction 请求 AI 玩家做出决定。
func (ge *GameEngine) RequestAIAction(playerID int32) {
	player := ge.getPlayerByID(playerID)
	if player == nil || !player.IsLlm { // TODO: 使此检查更通用（例如，IsAI）
		return
	}

	ge.logger.Info("Triggering AI for player", zap.String("player", player.Name))

	// 为动作生成器创建上下文
	ctx := &ai.ActionGeneratorContext{
		Player:        player,
		PlayerView:    ge.GeneratePlayerView(playerID),
		AllCharacters: ge.GameState.Characters,
	}

	go func() {
		action, err := ge.actionGenerator.GenerateAction(context.Background(), ctx)
		if err != nil {
			ge.logger.Error("AI action generation failed", zap.String("player", player.Name), zap.Error(err))
			// 提交默认操作以解锁游戏
			ge.engineChan <- &actionCompleteRequest{
				playerID: playerID,
				action:   &model.PlayerActionPayload{},
			}
			return
		}

		// 将经过验证的操作发送回主循环进行处理。
		ge.engineChan <- &actionCompleteRequest{
			playerID: playerID,
			action:   action,
		}
	}()
}

// GeneratePlayerView 为特定玩家创建游戏状态的过滤视图。
// 此方法不是线程安全的，必须仅在 runGameLoop goroutine 中调用。
func (ge *GameEngine) GeneratePlayerView(playerID int32) *model.PlayerView {
	player := ge.GameState.Players[playerID]
	if player == nil {
		return &model.PlayerView{}
	}

	view := &model.PlayerView{
		GameId:             ge.GameState.GameId,
		CurrentDay:         ge.GameState.CurrentDay,
		CurrentLoop:        ge.GameState.CurrentLoop,
		CurrentPhase:       ge.GameState.CurrentPhase,
		ActiveTragedies:    ge.GameState.ActiveTragedies,
		PreventedTragedies: ge.GameState.PreventedTragedies,
		PublicEvents:       ge.GameState.DayEvents,
	}

	// Filter character information based on player role
	view.Characters = make(map[int32]*model.PlayerViewCharacter, len(ge.GameState.Characters))
	for id, char := range ge.GameState.Characters {
		playerViewChar := &model.PlayerViewCharacter{
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
		if player.Role == model.PlayerRole_PLAYER_ROLE_PROTAGONIST {
			// Hide the true role from protagonists, show as unknown.
			playerViewChar.Role = model.RoleType_ROLE_TYPE_ROLE_UNKNOWN
		} else {
			playerViewChar.Role = char.HiddenRole
		}
		view.Characters[id] = playerViewChar
	}

	// Filter player information
	view.Players = make(map[int32]*model.PlayerViewPlayer, len(ge.GameState.Players))
	for id, p := range ge.GameState.Players {
		view.Players[id] = &model.PlayerViewPlayer{
			Id:   id,
			Name: p.Name,
			Role: p.Role,
		}
	}

	// Add player-specific information
	view.YourHand = player.Hand.Cards
	if player.Role == model.PlayerRole_PLAYER_ROLE_PROTAGONIST {
		view.YourDeductions = player.DeductionKnowledge
	}

	return view
}
