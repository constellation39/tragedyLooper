package engine

import (
	"fmt"
	"tragedylooper/internal/game/engine/effecthandler"
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

type targetManager struct {
	engine *GameEngine
}

func newTargetManager(engine *GameEngine) *targetManager {
	return &targetManager{engine: engine}
}

func (tm *targetManager) ResolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, ctx *effecthandler.EffectContext) ([]int32, error) {
	return tm.resolveSelector(gs, sel, ctx)
}

func (tm *targetManager) resolveSelector(gs *model.GameState, sel *model.TargetSelector, ctx *effecthandler.EffectContext) ([]int32, error) {
	if sel == nil {
		return nil, fmt.Errorf("target selector is nil")
	}

	var characterIDs []int32

	switch sel.SelectorType {
	case model.TargetSelector_SELECTOR_TYPE_SPECIFIC_CHARACTER:
		characterIDs = append(characterIDs, sel.CharacterId)
	case model.TargetSelector_SELECTOR_TYPE_ALL_CHARACTERS_AT_LOCATION:
		characterIDs = tm.engine.cm.GetCharactersInLocation(sel.LocationFilter)
	case model.TargetSelector_SELECTOR_TYPE_ALL_CHARACTERS:
		characterIDs = tm.engine.cm.GetAllCharacterIDs()
	case model.TargetSelector_SELECTOR_TYPE_ABILITY_USER:
		if ctx != nil && ctx.Payload != nil {
			characterIDs = append(characterIDs, ctx.Payload.PlayerId)
		}
	case model.TargetSelector_SELECTOR_TYPE_ABILITY_TARGET:
		if ctx != nil && ctx.Payload != nil {
			if t, ok := ctx.Payload.Target.(*model.UseAbilityPayload_TargetCharacterId); ok {
				characterIDs = append(characterIDs, t.TargetCharacterId)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported target selector type: %v", sel.SelectorType)
	}

	// TODO: Apply filters from the selector
	return characterIDs, nil
}
