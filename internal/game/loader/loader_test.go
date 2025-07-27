
package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadScript(t *testing.T) {
	script, err := LoadScript("../../../data/scripts/the_dark_forest.json")

	assert.NoError(t, err)
	assert.NotNil(t, script)

	if script != nil {
		assert.Equal(t, "The Dark Forest", script.Name)
		assert.Equal(t, int32(5), script.LoopCount)
	}
}
