package target

import (
	"fmt"
	"github.com/constellation39/tragedyLooper/internal/game/engine/effecthandler"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// ResolveSelectorToCharacters resolves a target selector to a list of character IDs.
func ResolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, ctx *effecthandler.EffectContext) ([]int32, error) {
	if sel == nil {
		return nil, fmt.Errorf("target selector is nil")
	}

	var characterIDs []int32

	switch sel.SelectorType {
	case model.TargetSelector_SELECTOR_TYPE_SPECIFIC_CHARACTER:
		characterIDs = append(characterIDs, sel.CharacterId)
	case model.TargetSelector_SELECTOR_TYPE_ALL_CHARACTERS_AT_LOCATION:
		characterIDs = getCharactersInLocation(gs, sel.LocationFilter)
	case model.TargetSelector_SELECTOR_TYPE_ALL_CHARACTERS:
		characterIDs = getAllCharacterIDs(gs)
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

func getCharactersInLocation(gs *model.GameState, location model.LocationType) []int32 {
	var charIDs []int32
	for id, char := range gs.Characters {
		if char.CurrentLocation == location {
			charIDs = append(charIDs, id)
		}
	}
	return charIDs
}

func getAllCharacterIDs(gs *model.GameState) []int32 {
	charIDs := make([]int32, 0, len(gs.Characters))
	for id := range gs.Characters {
		charIDs = append(charIDs, id)
	}
	return charIDs
}
