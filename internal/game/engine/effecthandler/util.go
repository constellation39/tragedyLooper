package effecthandler

import (
	"fmt"
	model "tragedylooper/pkg/proto/v1"
)

// CreateChoicesFromSelector is a helper function to generate player choices if a target selector
// resolves to more than one character.
func CreateChoicesFromSelector(ge GameEngine, selector *model.TargetSelector, ctx *EffectContext, description string) ([]*model.Choice, error) {
	state := ge.GetGameState()
	// We pass nils here because we are just trying to find out *if* a choice is needed.
	charIDs, err := ge.ResolveSelectorToCharacters(state, selector, ctx)
	if err != nil {
		return nil, err
	}

	// If the selector resolves to more than one character, a choice is required.
	if len(charIDs) > 1 {
		var choices []*model.Choice
		for _, charID := range charIDs {
			char, ok := state.Characters[charID]
			if !ok {
				continue
			}
			choiceID := fmt.Sprintf("target_char_%d", charID)
			choices = append(choices, &model.Choice{
				Id:          choiceID,
				Description: fmt.Sprintf("%s: %s", description, char.Config.Name),
				ChoiceType:  &model.Choice_TargetCharacterId{TargetCharacterId: charID},
			})
		}
		return choices, nil
	}

	return nil, nil
}
