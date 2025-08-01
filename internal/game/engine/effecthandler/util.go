package effecthandler

import (
	"fmt"

	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// CreateChoicesFromSelector 是一个辅助函数，用于在目标选择器
// 解析为多个角色时生成玩家选项。
func CreateChoicesFromSelector(ge GameEngine, selector *model.TargetSelector, ctx *EffectContext, description string) ([]*model.Choice, error) {
	state := ge.GetGameState()
	// 我们在这里传递 nil，因为我们只是想知道是否需要一个选择。
	charIDs, err := ge.ResolveSelectorToCharacters(state, selector, ctx)
	if err != nil {
		return nil, err
	}

	// 如果选择器解析为多个角色，则需要一个选择。
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
