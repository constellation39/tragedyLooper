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
	return tm.resolveSelector(gs, sel, ctx)
}

func (tm *TargetManager) resolveSelector(gs *model.GameState, sel *model.TargetSelector, ctx *effecthandler.EffectContext) ([]int32, error) {
	if sel == nil {
		return nil, fmt.Errorf("target selector is nil")
	}

	var characterIDs []int32
	var err error

	switch s := sel.SelectorType; {
	case s == model.TargetSelector_SPECIFIC_CHARACTER:
		characterIDs = append(characterIDs, sel.CharacterId)
	case s == model.TargetSelector_ALL_CHARACTERS_AT_LOCATION:
		characterIDs = tm.engine.cm.GetCharactersInLocation(sel.LocationFilter)
	case s == model.TargetSelector_ALL_CHARACTERS:
		characterIDs = tm.engine.cm.GetAllCharacterIDs()
	case s == model.TargetSelector_ABILITY_USER:
		if ctx != nil && ctx.Payload != nil {
			characterIDs = append(characterIDs, ctx.Payload.PlayerId)
		}
	case s == model.TargetSelector_ABILITY_TARGET:
		if ctx != nil && ctx.Payload != nil {
			if t, ok := ctx.Payload.Target.(*model.UseAbilityPayload_TargetCharacterId); ok {
				characterIDs = append(characterIDs, t.TargetCharacterId)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported target selector type: %T", s)
	}

	if err != nil {
		return nil, err
	}

	// TODO: apply filters from selector
	return characterIDs, nil
}
