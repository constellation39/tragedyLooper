package engine

import (
	"context"
	"fmt"
	"tragedylooper/internal/game/engine/effecthandler"
	"tragedylooper/internal/game/engine/eventhandler"
	"tragedylooper/internal/game/engine/phase"
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LocationGrid 定义了 2x2 的地图布局
var LocationGrid = map[model.LocationType]struct{ X, Y int }{
	model.LocationType_LOCATION_TYPE_SHRINE:   {0, 0},
	model.LocationType_LOCATION_TYPE_SCHOOL:   {1, 0},
	model.LocationType_LOCATION_TYPE_HOSPITAL: {0, 1},
	model.LocationType_LOCATION_TYPE_CITY:     {1, 1},
}

// engineAction 是一个空接口，用于标记所有可以发送到游戏引擎主循环的请求类型。
type engineAction interface{}

// getPlayerViewRequest 是获取玩家过滤后的游戏状态视图的请求。
type getPlayerViewRequest struct {
	playerID     int32
	responseChan chan *model.PlayerView
}

// actionCompleteRequest 表示 AI 或玩家操作已完成并准备好由游戏引擎处理。
type actionCompleteRequest struct {
	playerID int32
	action   *model.PlayerActionPayload
}

// getCurrentPhaseRequest 是一个安全获取当前游戏阶段的请求。
type getCurrentPhaseRequest struct {
	responseChan chan model.GamePhase
}

// GameEngine 管理单个游戏实例的状态和逻辑。
type GameEngine struct {
	GameState *model.GameState
	logger    *zap.Logger

	actionGenerator ActionGenerator
	gameConfig      loader.GameConfig
	pm              *phase.PhaseManager
	em              *eventhandler.Manager
	im              *incidentManager
	cm              *characterManager
	cc              *conditionChecker
	tm              *targetManager

	// engineChan 是所有传入请求（玩家操作、AI 操作等）的中央通道。
	engineChan chan engineAction
	stopChan   chan struct{}

	playerReady map[int32]bool

	mastermindPlayerID   int32
	protagonistPlayerIDs []int32
}

// NewGameEngine 创建一个新的游戏引擎实例。
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
	ge.pm = phase.NewPhaseManager(ge)
	ge.em = eventhandler.NewManager(ge)
	ge.im = newIncidentManager(ge)
	ge.cm = newCharacterManager(ge)
	ge.cc = newConditionChecker(ge)
	ge.tm = newTargetManager(ge)

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

// StartGameLoop 启动游戏主循环。
func (ge *GameEngine) StartGameLoop() {
	go ge.runGameLoop()
}

// StopGameLoop 停止游戏主循环。
func (ge *GameEngine) StopGameLoop() {
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

	// 这将阻塞，直到主游戏循环处理请求并发送响应。
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
	defer ge.em.Close()

	for {
		select {
		case <-ge.stopChan:
			return
		case req := <-ge.engineChan:
			ge.handleEngineRequest(req)
		case <-ge.pm.Timer():
			ge.pm.HandleTimeout()
		}
	}
}

// handleEngineRequest 处理来自引擎通道的传入请求。
func (ge *GameEngine) handleEngineRequest(req engineAction) {
	switch r := req.(type) {
	case *actionCompleteRequest:
		// AI 或玩家已提交操作。
		player, ok := ge.GameState.Players[r.playerID]
		if !ok {
			ge.logger.Warn("Action from unknown player", zap.Int32("playerID", r.playerID))
			return
		}
		ge.pm.HandleAction(player, r.action)
	case *getPlayerViewRequest:
		// 对特定于玩家的游戏状态视图的请求。
		r.responseChan <- ge.GeneratePlayerView(r.playerID)
	case *getCurrentPhaseRequest:
		r.responseChan <- ge.pm.CurrentPhase().Type()
	default:
		ge.logger.Warn("Unhandled request type in engine channel")
	}
}

func (ge *GameEngine) ApplyAndPublishEvent(eventType model.GameEventType, payload *model.EventPayload) {
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

	// Step 2: Let the current phase react to the event.
	ge.pm.HandleEvent(event)

	// Step 3: Record the event in the game state for player review.
	gs := ge.GetGameState()
	gs.DayEvents = append(gs.DayEvents, event)
	gs.LoopEvents = append(gs.LoopEvents, event)

	// Step 4: Publish the event to external listeners.
	ge.em.Dispatch(event)

	// Now that the state is updated, check if this event triggered any incidents.
	ge.im.TriggerIncidents()
}

// checkForTriggers is called after any state-changing event. It iterates through
// all incidents and checks if their trigger conditions are now met.
func (ge *GameEngine) checkForTriggers(event *model.GameEvent) {
	// TODO: We could optimize this by mapping event types to potentially affected incidents.
	// For now, we check all incidents.
	ge.im.TriggerIncidents()
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

func (ge *GameEngine) TriggerIncidents() {
	ge.im.TriggerIncidents()
}

func (ge *GameEngine) MoveCharacter(char *model.Character, dx, dy int) {
	ge.cm.MoveCharacter(char, dx, dy)
}

func (ge *GameEngine) CheckCondition(condition *model.Condition) (bool, error) {
	return ge.cc.Check(ge.GameState, condition)
}

func (ge *GameEngine) ResolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, ctx *effecthandler.EffectContext) ([]int32, error) {
	return ge.tm.ResolveSelectorToCharacters(gs, sel, ctx)
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
		ge.ApplyAndPublishEvent(model.GameEventType_GAME_EVENT_TYPE_CHOICE_REQUIRED, &model.EventPayload{
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

// TriggerAIPlayerAction 提示 AI 玩家做出决定。
func (ge *GameEngine) TriggerAIPlayerAction(playerID int32) {
	player := ge.getPlayerByID(playerID)
	if player == nil || !player.IsLlm { // TODO: 使此检查更通用（例如，IsAI）
		return
	}

	ge.logger.Info("Triggering AI for player", zap.String("player", player.Name))

	// 为动作生成器创建上下文
	ctx := &ActionGeneratorContext{
		Player:        player,
		PlayerView:    ge.GetPlayerView(playerID),
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
