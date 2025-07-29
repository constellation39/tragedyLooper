package engine

import (
	"context"
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// LocationGrid 定义了 2x2 的地图布局。
var LocationGrid = map[model.LocationType]struct{ X, Y int }{
	model.LocationType_SHRINE:   {0, 0},
	model.LocationType_SCHOOL:   {1, 0},
	model.LocationType_HOSPITAL: {0, 1},
	model.LocationType_CITY:     {1, 1},
}

// GameEngine 管理单个游戏实例的状态和逻辑。
type engineAction interface{}

// getPlayerViewRequest 是获取玩家过滤后的游戏状态视图的请求。
type getPlayerViewRequest struct {
	playerID     int32
	responseChan chan *model.PlayerView
}

type aiActionCompleteRequest struct {
	playerID int32
	action   *model.PlayerActionPayload
}
type GameEngine struct {
	GameState *model.GameState
	logger    *zap.Logger

	actionGenerator ActionGenerator
	gameConfig      loader.GameConfig
	pm              *phaseManager
	em              *eventManager

	// engineChan 是所有传入请求（玩家操作、AI 操作等）的中央通道。
	// 它确保对游戏状态的所有修改都在主游戏循环中按顺序处理，
	// 防止竞争条件。
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

	ge.initializeGameStateFromScript(gameConfig, playerMap)
	ge.dealInitialCards()

	return ge, nil
}

func (ge *GameEngine) StartGameLoop() {
	go ge.runGameLoop()
}

func (ge *GameEngine) StopGameLoop() {
	close(ge.stopChan)
}

func (ge *GameEngine) SubmitPlayerAction(playerID int32, action *model.PlayerActionPayload) {
	if action == nil {
		ge.logger.Warn("Received nil action from player")
		return
	}
	select {
	case ge.engineChan <- &aiActionCompleteRequest{playerID: playerID, action: action}:
	default:
		ge.logger.Warn("Request channel full, dropping action", zap.Int32("playerID", playerID))
	}
}

func (ge *GameEngine) GetGameEvents() <-chan *model.GameEvent {
	return ge.em.eventsChannel()
}

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
func (ge *GameEngine) handleEngineRequest(req engineAction) {
	switch r := req.(type) {
	case *aiActionCompleteRequest:
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

func (ge *GameEngine) ApplyAndPublishEvent(eventType model.GameEventType, payload proto.Message) {
	ge.em.createAndProcess(eventType, payload)
}

func (ge *GameEngine) endGame(winner model.PlayerRole) {
	ge.logger.Info("Game over", zap.String("winner", winner.String()))
	// 此事件将由事件管理器处理，从而导致状态更新和阶段转换。
	ge.em.createAndProcess(model.GameEventType_GAME_ENDED, &model.GameOverEvent{Winner: winner})
}

func (ge *GameEngine) ResetPlayerReadiness() {
	for playerID := range ge.GameState.Players {
		ge.playerReady[playerID] = false
	}
}

func (ge *GameEngine) GetCharacterByID(charID int32) *model.Character {
	char, ok := ge.GameState.Characters[charID]
	if !ok {
		return nil
	}
	return char
}

func (ge *GameEngine) TriggerIncidents() {
	// TODO: 实现事件触发逻辑
}

func (ge *GameEngine) MoveCharacter(char *model.Character, dx, dy int) {
	ge.moveCharacter(char, dx, dy)
}

func (ge *GameEngine) getPlayerByID(playerID int32) *model.Player {
	player, ok := ge.GameState.Players[playerID]
	if !ok {
		return nil
	}
	return player
}

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

		char.CurrentLocation = newLoc
		ge.ApplyAndPublishEvent(model.GameEventType_CHARACTER_MOVED, &model.CharacterMovedEvent{
			CharacterId: char.Config.Id,
			NewLocation: newLoc,
		})
		ge.logger.Info("character moved", zap.String("char", char.Config.Name), zap.String("to", newLoc.String()))
	}
}

// GetGameState 实现 phases.GameEngine 接口。
func (ge *GameEngine) GetGameState() *model.GameState {
	return ge.GameState
}

func (ge *GameEngine) GetGameConfig() loader.GameConfig {
	return ge.gameConfig
}

func (ge *GameEngine) AreAllPlayersReady() bool {
	// TODO: 实现我
	return false
}

func (ge *GameEngine) Logger() *zap.Logger {
	return ge.logger
}

func (ge *GameEngine) SetPlayerReady(playerID int32) {
	ge.playerReady[playerID] = true
}

func (ge *GameEngine) ResolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, player *model.Player, payload *model.UseAbilityPayload, ability *model.Ability) ([]int32, error) {
	return []int32{}, nil
}

// --- AI 集成 ---

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
		Script:        ge.gameConfig.GetScript(),
		AllCharacters: ge.GameState.Characters,
	}

	go func() {
		action, err := ge.actionGenerator.GenerateAction(context.Background(), ctx)
		if err != nil {
			ge.logger.Error("AI action generation failed", zap.String("player", player.Name), zap.Error(err))
			// 提交默认操作以解锁游戏
			ge.engineChan <- &aiActionCompleteRequest{
				playerID: playerID,
				action:   &model.PlayerActionPayload{},
			}
			return
		}

		// 将经过验证的操作发送回主循环进行处理。
		ge.engineChan <- &aiActionCompleteRequest{
			playerID: playerID,
			action:   action,
		}
	}()
}