package engine // 定义游戏引擎包

import (
	"context"                            // 导入 context 包，用于管理请求的生命周期
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

// GameEngine manages the state and logic of a single game instance.
type GameEngine struct {
	GameState *model.GameState // The current state of the game.
	logger    *zap.Logger      // Logger for logging.

	actionGenerator ActionGenerator         // Interface for generating actions for AI players.
	gameConfig      loader.GameDataAccessor // The data repository for the game.
	pm              *phaseManager           // Phase manager.
	em              *eventManager           // Event manager.

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
func NewGameEngine(logger *zap.Logger, players []*model.Player, actionGenerator ActionGenerator, gameConfig loader.GameDataAccessor) (*GameEngine, error) {
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

	ge.initializeGameStateFromScript(playerMap)
	ge.dealInitialCards()

	return ge, nil
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
		ge.pm.handleAction(r.playerID, r.action)
	case *getPlayerViewRequest:
		// 对特定于玩家的游戏状态视图的请求。
		r.responseChan <- ge.GeneratePlayerView(r.playerID)
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

// TriggerIncidents 触发事件。
func (ge *GameEngine) TriggerIncidents() {
	// TODO: 实现事件触发逻辑
}

// MoveCharacter 移动角色。
// char: 要移动的角色。
// dx: X轴上的移动量。
// dy: Y轴上的移动量。
func (ge *GameEngine) MoveCharacter(char *model.Character, dx, dy int) {
	ge.moveCharacter(char, dx, dy)
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

// moveCharacter 移动角色到新位置。
// char: 要移动的角色。
// dx: X轴上的移动量。
// dy: Y轴上的移动量。
func (ge *GameEngine) moveCharacter(char *model.Character, dx, dy int) {
	startPos, ok := LocationGrid[char.CurrentLocation]
	if !ok {
		ge.logger.Warn("character in unknown location", zap.String("char", char.Config.Name))
		return
	}

	// 计算新位置，在 2x2 网格上环绕。
	newX := (startPos.X + dx) % 2
	newY := (startPos.Y + dy) % 2

	var newLoc model.LocationType
	for loc, pos := range LocationGrid {
		if pos.X == newX && pos.Y == newY {
			newLoc = loc
			break
		}
	}

	if newLoc != model.LocationType_LOCATION_TYPE_UNSPECIFIED && newLoc != char.CurrentLocation {
		// 检查移动限制
		for _, rule := range char.Config.Rules {
			if smr, ok := rule.Effect.(*model.CharacterRule_SpecialMovementRule); ok {
				for _, restricted := range smr.SpecialMovementRule.RestrictedLocations {
					if restricted == newLoc {
						ge.logger.Info("character movement restricted", zap.String("char", char.Config.Name), zap.String("location", newLoc.String()))
						return // 禁止移动
					}
				}
			}
		}

		ge.ApplyAndPublishEvent(model.GameEventType_CHARACTER_MOVED, &model.EventPayload{
			Payload: &model.EventPayload_CharacterMoved{CharacterMoved: &model.CharacterMovedEvent{
				CharacterId: char.Config.Id,
				NewLocation: newLoc,
			}},
		})
	}
}

// GetGameState 实现 phases.GameEngine 接口，返回当前游戏状态。
func (ge *GameEngine) GetGameState() *model.GameState {
	return ge.GameState
}

func (ge *GameEngine) GetGameRepo() loader.GameDataAccessor {
	return ge.gameConfig
}

// AreAllPlayersReady 检查所有玩家是否都已准备好。
// TODO: 实现我
func (ge *GameEngine) AreAllPlayersReady() bool {
	return false
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

// ResolveSelectorToCharacters 根据目标选择器解析出对应的角色ID列表。
// gs: 游戏状态。
// sel: 目标选择器。
// player: 相关的玩家（如果适用）。
// payload: 相关的操作负载（如果适用）。
// ability: 相关的能力（如果适用）。
// 返回值: 角色ID列表和可能发生的错误。
func (ge *GameEngine) ResolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, player *model.Player, payload *model.UseAbilityPayload, ability *model.Ability) ([]int32, error) {
	return []int32{}, nil
}

// --- AI 集成 ---

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
