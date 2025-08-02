package phasehandler

import (
	"github.com/constellation39/tragedyLooper/internal/game/loader"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// GameEngine defines the interface that phases can interact with the game engine.
type GameEngine interface {
	// TriggerEvent applies and publishes an event.
	TriggerEvent(eventType model.GameEventType, eventData *model.EventPayload)
	// Logger returns the logger for the game engine.
	Logger() *zap.Logger
	// GetGameState returns the current game state.
	GetGameState() *model.GameState
	// GetGameRepo returns the game configuration.
	GetGameRepo() loader.GameConfig
	// GetCharacterByID retrieves a character by their ID.
	GetCharacterByID(id int32) *model.Character
	// MoveCharacter moves a character.
	MoveCharacter(char *model.Character, dx, dy int)
	GetMastermindPlayer() *model.Player
	GetProtagonistPlayers() []*model.Player
	ApplyEffect(effect *model.Effect, ability *model.Ability, payload *model.UseAbilityPayload, choice *model.ChooseOptionPayload) error
	RequestAIAction(playerID int32)
}

// Phase is the interface for a game phase, defining methods that each phase must implement.
type Phase interface {
	// Type returns the type of the phase.
	Type() model.GamePhase
	// Enter is called when entering this phase.
	Enter(ge GameEngine)
	// HandleAction handles a player's action in this phase.
	HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) bool
	// HandleEvent handles a game event received in this phase.
	HandleEvent(ge GameEngine, event *model.GameEvent) bool
	// HandleTimeout handles a timeout in this phase.
	HandleTimeout(ge GameEngine)
	// isReadyToTransition checks if the phase is ready to transition to the next phase.
	isReadyToTransition() bool
	// Exit is called when exiting this phase.
	Exit(ge GameEngine)
	// TimeoutTicks returns the timeout duration in game ticks for this phase.
	TimeoutTicks() int64
}
