package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadGameData(t *testing.T) {
	gameData, err := LoadGameData("../../../data")

	assert.NoError(t, err)
	assert.NotNil(t, gameData)

	if gameData != nil {
		assert.NotEmpty(t, gameData.Scripts)
		script, ok := gameData.Scripts["the_dark_forest.json"]
		assert.True(t, ok)
		assert.NotNil(t, script)
		if script != nil {
			assert.Equal(t, "The Dark Forest", script.Name)
			assert.Equal(t, int32(5), script.LoopCount)
		}
	}
}
