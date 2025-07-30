package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register(model.GameEventType_TRAIT_ADDED, &TraitAddedHandler{})
}

// TraitAddedHandler 处理 TraitAddedEvent。
type TraitAddedHandler struct{}

// Handle 如果特征尚不存在，则将其添加到角色中。
func (h *TraitAddedHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	e, ok := event.Payload.Payload.(*model.EventPayload_TraitAdded)
	if !ok {
		return nil // 或适当处理错误
	}

	state := ge.GetGameState()
	if char, ok := state.Characters[e.TraitAdded.CharacterId]; ok {
		// 避免重复
		for _, t := range char.Traits {
			if t == e.TraitAdded.Trait {
				return nil // 已存在，不是错误
			}
		}
		char.Traits = append(char.Traits, e.TraitAdded.Trait)
	}
	return nil
}