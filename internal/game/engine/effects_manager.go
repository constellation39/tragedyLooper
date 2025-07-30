package engine

import (
	"tragedylooper/internal/game/engine/effecthandler"
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// GetEffectDescription 查找适当的处理程序并返回效果的描述。
func (ge *GameEngine) GetEffectDescription(effect *model.Effect) string {
	return effecthandler.GetEffectDescription(ge, effect)
}
