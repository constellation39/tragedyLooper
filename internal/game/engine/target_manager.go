package engine

import (
	"fmt"
	"tragedylooper/internal/game/engine/effecthandler"
	model "tragedylooper/pkg/proto/v1"
)

type TargetManager struct {
	engine *GameEngine
}

func NewTargetManager(engine *GameEngine) *TargetManager {
	return &TargetManager{engine: engine}
}

func (tm *TargetManager) ResolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, ctx *effecthandler.EffectContext) ([]int32, error) {
	// TODO: Implement this method based on the game's logic for target selection.
	return nil, fmt.Errorf("ResolveSelectorToCharacters not implemented")
}
