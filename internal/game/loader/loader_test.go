package loader

import (
	"path/filepath"
	"testing"

	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Use the actual data directory
	dataDir, err := filepath.Abs("../../../data")
	assert.NoError(t, err)

	// Load the config for the "first_steps" script
	config, err := LoadConfig(dataDir, "basic_tragedy_x", 101)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Check script
	script := config.GetScript()
	assert.NotNil(t, script)
	assert.Equal(t, "Basic Tragedy Set X", script.Name)
	// assert.Equal(t, int32(3), script.LoopCount)
	// assert.Equal(t, int32(4), script.DaysPerLoop)

	// Check abilities
	abilities := config.GetAbilityMap()
	assert.NotEmpty(t, abilities)

	// Check cards
	cards := config.GetCardMap()
	assert.Len(t, cards, 15)
	card1 := cards[6001]
	assert.Equal(t, "Move", card1.Name)
	assert.Equal(t, v1.CardType_CARD_TYPE_MOVE_HORIZONTALLY, card1.GetCardType())
	assert.Equal(t, v1.PlayerRole_PLAYER_ROLE_MASTERMIND, card1.OwnerRole)

	// Check characters
	characters := config.GetCharacterMap()
	assert.NotEmpty(t, characters)

	// Check incidents
	incidents := config.GetIncidentMap()
	assert.NotEmpty(t, incidents)
}
