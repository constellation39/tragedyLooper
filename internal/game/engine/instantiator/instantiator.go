// Package instantiator handles the creation of the initial game state from script configurations.
package instantiator

import (
	"github.com/constellation39/tragedyLooper/internal/game/loader"
	pb "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
	"github.com/google/uuid"
)

// NewGameState converts a ScriptConfig and ScriptModel protobuf message to a GameState message.
func NewGameState(players []*pb.Player, gameConfig loader.ScriptConfig) *pb.GameState {
	script := gameConfig.GetScript()
	model := gameConfig.GetModel()

	if model == nil || script == nil {
		return nil
	}

	characters := make(map[int32]*pb.Character, len(model.PrivateConfig.CharactersIds))
	for _, charID := range model.PrivateConfig.CharactersIds {
		charConfig, ok := script.Characters[charID]
		if !ok {
			continue
		}
		roleID, ok := model.PrivateConfig.RoleAssignments[charID]
		if !ok {
			roleID = 0 // Assuming 0 is an invalid/unknown role ID
		}
		characters[charID] = newCharacterFromConfig(charConfig, roleID)
	}

	for _, player := range players {
		switch player.Role {
		case pb.PlayerRole_PLAYER_ROLE_PROTAGONIST:
			player.Hand = &pb.CardList{Cards: newCardsFromConfig(script.ProtagonistCards)}
		case pb.PlayerRole_PLAYER_ROLE_MASTERMIND:
			player.Hand = &pb.CardList{Cards: newCardsFromConfig(script.MastermindCards)}
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

// newCardsFromConfig converts a map of CardConfig protos to a slice of Card runtime instances.
func newCardsFromConfig(configs map[int32]*pb.CardConfig) []*pb.Card {
	cards := make([]*pb.Card, 0, len(configs))
	for _, cardConfig := range configs {
		cards = append(cards, newCardFromConfig(cardConfig))
	}
	return cards
}

// newCharacterFromConfig converts a CharacterConfig protobuf message to a Character runtime instance.
func newCharacterFromConfig(config *pb.CharacterConfig, roleId int32) *pb.Character {
	if config == nil {
		return nil
	}

	abilities := make([]*pb.Ability, len(config.Abilities))
	for i, abilityConfig := range config.Abilities {
		abilities[i] = newAbilityFromConfig(abilityConfig)
	}

	stats := make(map[int32]int32)
	if config.StatLimits != nil {
		for stat := range config.StatLimits {
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

// newIncidentFromConfig converts an IncidentConfig protobuf message to an Incident message.
func newIncidentFromConfig(config *pb.IncidentConfig) *pb.Incident {
	if config == nil {
		return nil
	}

	return &pb.Incident{
		Config:               config,
		HasTriggeredThisLoop: false,
	}
}

// newCardFromConfig converts a CardConfig protobuf message to a Card message.
func newCardFromConfig(config *pb.CardConfig) *pb.Card {
	if config == nil {
		return nil
	}

	return &pb.Card{
		Config:         config,
		UsedThisLoop:   false,
		ResolvedTarget: nil,
	}
}

// newAbilityFromConfig converts an AbilityConfig protobuf message to an Ability message.
func newAbilityFromConfig(config *pb.AbilityConfig) *pb.Ability {
	if config == nil {
		return nil
	}

	return &pb.Ability{
		Config:           config,
		UsedThisLoop:     false,
		OwnerCharacterId: 0,
	}
}
