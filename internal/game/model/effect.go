package model

import (
	"encoding/json"
	"fmt"
)

// Effect 是所有能力效果都必须实现的接口。
// 它使用策略模式，将每种效果的逻辑封装到自己的类型中。
type Effect interface {
	// Apply 将效果应用到游戏状态。
	// 它接收一个 GameMutator 接口来修改状态，以及一个包含上下文的 payload。
	Apply(mutator GameMutator, payload UseAbilityPayload) error
}

// --- 具体效果实现 ---

// MoveCharacterEffect 移动一个角色到指定地点。
type MoveCharacterEffect struct{}

func (e *MoveCharacterEffect) Apply(mutator GameMutator, payload UseAbilityPayload) error {
	if payload.TargetCharacterID == "" {
		return fmt.Errorf("MoveCharacterEffect requires a TargetCharacterID")
	}
	if payload.TargetLocation == "" {
		return fmt.Errorf("MoveCharacterEffect requires a TargetLocation")
	}
	mutator.SetCharacterLocation(payload.TargetCharacterID, payload.TargetLocation)
	return nil
}

// AdjustParanoiaEffect 调整角色的妄想值。
type AdjustParanoiaEffect struct {
	Amount int `json:"amount"`
}

func (e *AdjustParanoiaEffect) Apply(mutator GameMutator, payload UseAbilityPayload) error {
	if payload.TargetCharacterID == "" {
		return fmt.Errorf("AdjustParanoiaEffect requires a TargetCharacterID")
	}
	mutator.AdjustCharacterParanoia(payload.TargetCharacterID, e.Amount)
	return nil
}

// AdjustGoodwillEffect 调整角色的好感度。
type AdjustGoodwillEffect struct {
	Amount int `json:"amount"`
}

func (e *AdjustGoodwillEffect) Apply(mutator GameMutator, payload UseAbilityPayload) error {
	if payload.TargetCharacterID == "" {
		return fmt.Errorf("AdjustGoodwillEffect requires a TargetCharacterID")
	}
	mutator.AdjustCharacterGoodwill(payload.TargetCharacterID, e.Amount)
	return nil
}

// AdjustIntrigueEffect 调整角色的阴谋值。
type AdjustIntrigueEffect struct {
	Amount int `json:"amount"`
}

func (e *AdjustIntrigueEffect) Apply(mutator GameMutator, payload UseAbilityPayload) error {
	if payload.TargetCharacterID == "" {
		return fmt.Errorf("AdjustIntrigueEffect requires a TargetCharacterID")
	}
	mutator.AdjustCharacterIntrigue(payload.TargetCharacterID, e.Amount)
	return nil
}

// --- 自定义反序列化 ---

// effectWrapper 是一个辅助结构体，用于在反序列化时识别效果类型。
type effectWrapper struct {
	Type   EffectType      `json:"type"`
	Params json.RawMessage `json:"params"`
}

// UnmarshalEffect 是一个辅助函数，用于将 JSON 数据反序列化为正确的 Effect 类型。
func UnmarshalEffect(data []byte) (Effect, error) {
	var wrapper effectWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal effect wrapper: %w", err)
	}

	var effect Effect
	switch wrapper.Type {
	case EffectTypeMoveCharacter:
		effect = &MoveCharacterEffect{}
	case EffectTypeAdjustParanoia:
		effect = &AdjustParanoiaEffect{}
	case EffectTypeAdjustGoodwill:
		effect = &AdjustGoodwillEffect{}
	case EffectTypeAdjustIntrigue:
		effect = &AdjustIntrigueEffect{}
	// TODO: 为其他效果类型添加 case
	default:
		return nil, fmt.Errorf("unknown effect type: '%s'", wrapper.Type)
	}

	// 将 params 部分反序列化到具体的 effect 结构体中
	if len(wrapper.Params) > 0 && string(wrapper.Params) != "null" {
		if err := json.Unmarshal(wrapper.Params, effect); err != nil {
			return nil, fmt.Errorf("failed to unmarshal params for type '%s': %w", wrapper.Type, err)
		}
	}

	return effect, nil
}