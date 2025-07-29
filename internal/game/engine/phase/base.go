package phase

import (
	"time"
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// GameEngine 定义了阶段可以与之交互的游戏引擎的接口。
type GameEngine interface {
	ApplyAndPublishEvent(eventType model.GameEventType, eventData proto.Message)
	AreAllPlayersReady() bool
	Logger() *zap.Logger
	ResetPlayerReadiness()
	SetPlayerReady(playerID int32)
	StopGameLoop()
	TriggerIncidents()
	GetGameState() *model.GameState
	GetGameConfig() loader.GameConfig
	GetCharacterByID(id int32) *model.Character
	MoveCharacter(char *model.Character, dx, dy int)
}

// Phase 是游戏阶段的接口。
type Phase interface {
	Type() model.GamePhase
	Enter(ge GameEngine) Phase
	HandleAction(ge GameEngine, playerID int32, action *model.PlayerActionPayload) Phase
	HandleEvent(ge GameEngine, event *model.GameEvent) Phase
	HandleTimeout(ge GameEngine) Phase
	Exit(ge GameEngine)
	TimeoutDuration() time.Duration
}

// basePhase 是一个辅助结构体，为 Phase 接口提供默认实现。
type basePhase struct{}

func (p *basePhase) Enter(ge GameEngine) Phase { return nil }
func (p *basePhase) HandleAction(ge GameEngine, playerID int32, action *model.PlayerActionPayload) Phase {
	return nil
}
func (p *basePhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }
func (p *basePhase) HandleTimeout(ge GameEngine) Phase                       { return nil }
func (p *basePhase) Exit(ge GameEngine)                                      {}
func (p *basePhase) TimeoutDuration() time.Duration                          { return 0 }