package phasehandler

import (
	"time"
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// GameEngine defines the interface that phases can interact with the game engine.
type GameEngine interface {
	// TriggerEvent applies and publishes an event.
	TriggerEvent(eventType model.GameEventType, eventData *model.EventPayload)
	// TriggerIncidents checks and triggers incidents.
	TriggerIncidents()
	// CheckCondition checks if a condition is met.
	CheckCondition(condition *model.Condition) (bool, error)
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
	// Enter is called when entering this phase, returns the next phase if it switches immediately.
	Enter(ge GameEngine) Phase
	// HandleAction handles a player's action in this phase, returns the next phase if it changes.
	HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase
	// HandleEvent handles a game event received in this phase, returns the next phase if it changes.
	HandleEvent(ge GameEngine, event *model.GameEvent) Phase
	// HandleTimeout handles a timeout in this phase, returns the next phase.
	HandleTimeout(ge GameEngine) Phase
	// Exit is called when exiting this phase.
	Exit(ge GameEngine)
	// TimeoutDuration returns the timeout duration for this phase.
	TimeoutDuration() time.Duration
}
