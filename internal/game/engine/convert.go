// Package engine handles game state initialization and management.
package engine

import (
	"github.com/constellation39/tragedyLooper/internal/game/loader"
	pb "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
	"github.com/google/uuid"
)

// InitializeGameStateFromScript converts a ScriptConfig and ScriptModel protobuf message to a GameState message.
func InitializeGameStateFromScript(players []*pb.Player, gameConfig loader.ScriptConfig) *pb.GameState {
	script := gameConfig.GetScript()
	model := gameConfig.GetModel()

	if model == nil || script == nil {
		return nil
	}

	characters := make(map[int32]*pb.Character, len(model.PrivateConfig.CharactersIds))
	for _, charID := range model.PrivateConfig.CharactersIds {
		charConfig, ok := script.Characters[charID]
		if !ok {
			// Handle error: character model not found for ID
			continue
		}
		roleID, ok := model.PrivateConfig.RoleAssignments[charID]
		if !ok {
			// Handle error: role assignment not found for character
			// Assign a default or unknown role
			roleID = 0 // Assuming 0 is an invalid/unknown role ID
		}
		characters[charID] = NewCharacterFromConfig(charConfig, roleID)
	}

	// Incidents, cards etc. would be initialized and held by the engine, not directly in the active GameState
	// The GameState holds the *current* state, not the library of all possible items.

	for _, player := range players {
		if player.Role == pb.PlayerRole_PLAYER_ROLE_MASTERMIND {

		}

		switch player.Role {
		case pb.PlayerRole_PLAYER_ROLE_PROTAGONIST:
			player.Hand = &pb.CardList{Cards: make([]*pb.Card, 0)}
			player.Hand.Cards = newCardsFromConfig(script.ProtagonistCards)
		case pb.PlayerRole_PLAYER_ROLE_MASTERMIND:
			player.Hand = &pb.CardList{Cards: make([]*pb.Card, 0)}
			player.Hand.Cards = newCardsFromConfig(script.MastermindCards)
		}

	}

	return &pb.GameState{
		GameId:             uuid.New().String(),
		Tick:               0,
		CurrentLoop:        1, // Game starts on Loop 1
		DaysPerLoop:        model.PublicConfig.DaysPerLoop,
		CurrentDay:         0, // Starts before Day 1 begins
		CurrentPhase:       pb.GamePhase_GAME_PHASE_SETUP,
		Characters:         characters,
		Players:            make(map[int32]*pb.Player), // Players will be added later
		TriggeredIncidents: make(map[int32]bool),
	}
}

func newCardsFromConfig(configs map[int32]*pb.CardConfig) []*pb.Card {
	cards := make([]*pb.Card, 0, len(configs))
	for _, cardConfig := range configs {
		cards = append(cards, NewCardFromConfig(cardConfig))
	}
	return cards
}

// NewCharacterFromConfig converts a CharacterConfig protobuf message to a Character runtime instance.
func NewCharacterFromConfig(config *pb.CharacterConfig, roleId int32) *pb.Character {
	if config == nil {
		return nil
	}

	abilities := make([]*pb.Ability, len(config.Abilities))
	for i, abilityConfig := range config.Abilities {
		abilities[i] = NewAbilityFromConfig(abilityConfig)
	}

	// Initialize stats map
	stats := make(map[int32]int32)
	// You might want to define constants for these stat types in your enums.proto
	// For now, let's assume Paranoia=1, Intrigue=2, Goodwill=3 based on common usage.
	// This part is a bit of a guess and should be standardized.
	// Let's assume CharacterConfig has initial values, if not, they are 0.
	// The current proto has no initial values in config, so let's check stat_limits
	if config.StatLimits != nil {
		for stat := range config.StatLimits {
			// This is also a guess, assuming stats start at 0 and we are just copying limits for now.
			// A better approach would be to have initial stats in the CharacterConfig.
			// Let's assume stats start at 0 unless specified.
			stats[stat] = 0
		}
	}

	return &pb.Character{
		Config:          config,
		CurrentLocation: config.InitialLocation,
		Stats:           stats,
		HiddenRoleId:    roleId,
		Abilities:       abilities,
		IsAlive:         true,
		InPanicMode:     false,
		Traits:          config.Traits, // Initial traits from config
	}
}

// NewIncidentFromConfig converts an IncidentConfig protobuf message to an Incident message.
func NewIncidentFromConfig(config *pb.IncidentConfig) *pb.Incident {
	if config == nil {
		return nil
	}

	return &pb.Incident{
		Config:               config,
		HasTriggeredThisLoop: false,
	}
}

// NewCardFromConfig converts a CardConfig protobuf message to a Card message.
func NewCardFromConfig(config *pb.CardConfig) *pb.Card {
	if config == nil {
		return nil
	}

	return &pb.Card{
		Config:         config,
		UsedThisLoop:   false,
		ResolvedTarget: nil,
	}
}

// NewAbilityFromConfig converts an AbilityConfig protobuf message to an Ability message.
func NewAbilityFromConfig(config *pb.AbilityConfig) *pb.Ability {
	if config == nil {
		return nil
	}

	return &pb.Ability{
		Config:           config,
		UsedThisLoop:     false,
		OwnerCharacterId: 0,
	}
}
