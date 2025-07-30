package engine // 定义游戏引擎包

import (
	"context" // 导入 context 包，用于管理请求的生命周期
	"fmt"
	"tragedylooper/internal/game/engine/effecthandler"
	"tragedylooper/internal/game/loader" // 导入游戏数据加载器
	model "tragedylooper/pkg/proto/v1"   // 导入协议缓冲区模型

	"go.uber.org/zap" // 导入 Zap 日志库
)

// LocationGrid 定义了 2x2 的地图布局，将地点类型映射到网格坐标。
var LocationGrid = map[model.LocationType]struct{ X, Y int }{
	model.LocationType_SHRINE:   {0, 0},
	model.LocationType_SCHOOL:   {1, 0},
	model.LocationType_HOSPITAL: {0, 1},
	model.LocationType_CITY:     {1, 1},
}

// engineAction 是一个空接口，用于标记所有可以发送到游戏引擎主循环的请求类型。
type engineAction interface{}

// getPlayerViewRequest 是获取玩家过滤后的游戏状态视图的请求。
type getPlayerViewRequest struct {
	playerID     int32                  // 请求视图的玩家ID
	responseChan chan *model.PlayerView // 用于发送响应的通道
}

// actionCompleteRequest 表示 AI 或玩家操作已完成并准备好由游戏引擎处理。
type actionCompleteRequest struct {
	playerID int32                      // 执行操作的玩家ID
	action   *model.PlayerActionPayload // 玩家操作的负载
}

// getCurrentPhaseRequest is a request to get the current game phase safely.
type getCurrentPhaseRequest struct {
	responseChan chan model.GamePhase
}

// GameEngine manages the state and logic of a single game instance.
type GameEngine struct {
	GameState *model.GameState // The current state of the game.
	logger    *zap.Logger      // Logger for logging.

	actionGenerator ActionGenerator   // Interface for generating actions for AI players.
	gameConfig      loader.GameConfig // The data repository for the game.
	pm              *phaseManager     // Phase manager.
	em              *eventManager     // Event manager.
	im              *IncidentManager
	cm              *CharacterManager
	cc              *ConditionChecker
	tm              *TargetManager

	// engineChan is the central channel for all incoming requests (player actions, AI actions, etc.).
	// It ensures that all modifications to the game state are processed sequentially in the main game loop,
	// preventing race conditions.
	engineChan chan engineAction // Channel for engine requests.
	stopChan   chan struct{}     // Channel to stop the game loop.

	playerReady map[int32]bool // Records whether each player is ready for the next phase.

	mastermindPlayerID   int32   // ID of the mastermind player.
	protagonistPlayerIDs []int32 // List of protagonist player IDs.
}

// NewGameEngine creates a new instance of the game engine.
// logger: The logger.
// players: The list of players in the game.
// actionGenerator: The AI action generator.
// gameConfig: The game configuration.
// Returns: A new GameEngine instance and a possible error.
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
	ge.im = NewIncidentManager(ge)
	ge.cm = NewCharacterManager(ge)
	ge.cc = NewConditionChecker(ge)
	ge.tm = NewTargetManager(ge)

	playerMap := ge.initializePlayers(players)

	ge.initializeGameStateFromScript(playerMap)
	ge.dealInitialCards()

	return ge, nil
}

func (ge *GameEngine) initializePlayers(players []*model.Player) map[int32]*model.Player {
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
	return playerMap
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
// playerID: 玩家ID。
// action: 玩家操作的负载。
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
	return ge.em.eventsChannel()
}

// GetPlayerView 获取指定玩家的游戏状态视图。
// playerID: 玩家ID。
// 返回值: 玩家的游戏视图。
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

// GetCurrentPhase safely gets the current game phase from the engine.
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

// handleEngineRequest 处理来自引擎通道的传入请求。
// req: 传入的引擎动作请求。
func (ge *GameEngine) handleEngineRequest(req engineAction) {
	switch r := req.(type) {
	case *actionCompleteRequest:
		// AI 或玩家已提交操作。
		player, ok := ge.GameState.Players[r.playerID]
		if !ok {
			ge.logger.Warn("Action from unknown player", zap.Int32("playerID", r.playerID))
			return
		}
		ge.pm.handleAction(player, r.action)
	case *getPlayerViewRequest:
		// 对特定于玩家的游戏状态视图的请求。
		r.responseChan <- ge.GeneratePlayerView(r.playerID)
	case *getCurrentPhaseRequest:
		r.responseChan <- ge.pm.CurrentPhase().Type()
	default:
		ge.logger.Warn("Unhandled request type in engine channel")
	}
}

// handleTimeout 在当前阶段的计时器到期时调用。
func (ge *GameEngine) handleTimeout() {
	ge.pm.handleTimeout()
}

func (ge *GameEngine) ApplyAndPublishEvent(eventType model.GameEventType, payload *model.EventPayload) {
	ge.em.createAndProcess(eventType, payload)
}

// endGame 结束游戏并宣布获胜方。
// winner: 获胜方的角色类型。
func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.logger.Info("Game over", zap.String("winner", winner.String()))
	// This event will be handled by the event manager, leading to state updates and phase transitions.
	ge.ApplyAndPublishEvent(model.GameEventType_GAME_ENDED, &model.EventPayload{Payload: &model.EventPayload_GameOver{GameOver: &model.GameOverEvent{Winner: winner}}})
}

// ResetPlayerReadiness 重置所有玩家的准备状态。
func (ge *GameEngine) ResetPlayerReadiness() {
	for playerID := range ge.GameState.Players {
		ge.playerReady[playerID] = false
	}
}

// GetCharacterByID 根据角色ID获取角色对象。
// charID: 角色ID。
// 返回值: 角色对象，如果未找到则返回 nil。
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
// playerID: 玩家ID。
// 返回值: 玩家对象，如果未找到则返回 nil。
func (ge *GameEngine) getPlayerByID(playerID int32) *model.Player {
	player, ok := ge.GameState.Players[playerID]
	if !ok {
		return nil
	}
	return player
}

// GetGameState 实现 phases.GameEngine 接口，返回当前游戏状态。
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
// playerID: 玩家ID。
func (ge *GameEngine) SetPlayerReady(playerID int32) {
	ge.playerReady[playerID] = true
}

// --- AI 集成 ---

// GetProtagonistPlayers returns the protagonist players.
func (ge *GameEngine) GetProtagonistPlayers() []*model.Player {
	players := make([]*model.Player, 0, len(ge.protagonistPlayerIDs))
	for _, id := range ge.protagonistPlayerIDs {
		players = append(players, ge.getPlayerByID(id))
	}
	return players
}

// GetMastermindPlayer returns the mastermind player.
func (ge *GameEngine) GetMastermindPlayer() *model.Player {
	return ge.getPlayerByID(ge.mastermindPlayerID)
}

// ApplyEffect finds the appropriate handler for an effect, resolves choices, and then applies the effect.
func (ge *GameEngine) ApplyEffect(effect *model.Effect, ability *model.Ability, payload interface{}, choice *model.ChooseOptionPayload) error {
	handler, err := effecthandler.GetEffectHandler(effect)
	if err != nil {
		return err
	}

	ctx := &effecthandler.EffectContext{
		Ability: ability,
		Payload: payload,
		Choice:  choice,
	}

	// 1. Resolve Choices
	choices, err := handler.ResolveChoices(ge, effect, ctx)
	if err != nil {
		return fmt.Errorf("error resolving choices: %w", err)
	}

	if len(choices) > 0 && choice == nil {
		choiceEvent := &model.ChoiceRequiredEvent{Choices: choices}
		ge.ApplyAndPublishEvent(model.GameEventType_CHOICE_REQUIRED, &model.EventPayload{
			Payload: &model.EventPayload_ChoiceRequired{ChoiceRequired: choiceEvent},
		})
		return nil // Stop processing until a choice is made
	}

	// 2. Apply Effect
	err = handler.Apply(ge, effect, ctx)
	if err != nil {
		return fmt.Errorf("error applying effect: %w", err)
	}

	return nil
}

// TriggerAIPlayerAction 提示 AI 玩家做出决定。
// playerID: AI 玩家的ID。
func (ge *GameEngine) TriggerAIPlayerAction(playerID int32) {
	player := ge.getPlayerByID(playerID)
	if player == nil || !player.IsLlm { // TODO: 使此检查更通用（例如，IsAI）
		return
	}

	ge.logger.Info("Triggering AI for player", zap.String("player", player.Name))

	// Create context for the action generator
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
