package phase

import (
	"time"
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// GameEngine defines the interface for the game engine that phases can interact with.
type GameEngine interface {
	ApplyAndPublishEvent(eventType model.GameEventType, eventData proto.Message)
	AreAllPlayersReady() bool
	Logger() *zap.Logger
	ResetPlayerReadiness()
	ResolveMovement()
	ResolveOtherCards()
	SetPlayerReady(playerID int32)
	StopGameLoop()
	TriggerIncidents()
	GetGameState() *model.GameState
	GetGameConfig() loader.GameConfig
}

// Phase is an interface for a game phase.
type Phase interface {
	Type() model.GamePhase
	Enter(ge GameEngine) Phase
	HandleAction(ge GameEngine, playerID int32, action *model.PlayerActionPayload) Phase
	HandleEvent(ge GameEngine, event *model.GameEvent) Phase
	HandleTimeout(ge GameEngine) Phase
	ResolveMovement(ge GameEngine) Phase
	ResolveOtherCards(ge GameEngine) Phase
	Exit(ge GameEngine)
	TimeoutDuration() time.Duration
}

// basePhase is a helper struct that provides default implementations for the Phase interface.
type basePhase struct{}

func (p *basePhase) Enter(ge GameEngine) Phase                                     { return nil }
func (p *basePhase) HandleAction(ge GameEngine, playerID int32, action *model.PlayerActionPayload) Phase { return nil }
func (p *basePhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase       { return nil }
func (p *basePhase) HandleTimeout(ge GameEngine) Phase                             { return nil }
func (p *basePhase) ResolveMovement(ge GameEngine) Phase                           { return nil }
func (p *basePhase) ResolveOtherCards(ge GameEngine) Phase                         { return nil }
func (p *basePhase) Exit(ge GameEngine)                                            {}
func (p *basePhase) TimeoutDuration() time.Duration                                { return 0 }
