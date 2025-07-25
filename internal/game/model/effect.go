package model

import (
	"encoding/json"
	"fmt"
)

// EffectContext 是一个由引擎创建的、只读的上下文，
// 包含了效果决策所需的所有信息。
type EffectContext struct {
	GameState *GameState
}

// TargetChoice 代表一个可供玩家选择的有效目标。
type TargetChoice struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// Effect 接口定义了一个效果的核心行为。
type Effect interface {
	ResolveChoices(ctx EffectContext, self *Ability) ([]TargetChoice, error)
	Execute(ctx EffectContext, self *Ability, payload UseAbilityPayload) ([]Event, error)
}

// --- 具体效果实现 ---

// MoveCharacterEffect 移动一个角色到指定地点。
type MoveCharacterEffect struct{}

func (e *MoveCharacterEffect) ResolveChoices(ctx EffectContext, self *Ability) ([]TargetChoice, error) {
	return nil, nil
}

func (e *MoveCharacterEffect) Execute(ctx EffectContext, self *Ability, payload UseAbilityPayload) ([]Event, error) {
	if payload.TargetCharacterID == "" || payload.TargetLocation == "" {
		return nil, fmt.Errorf("MoveCharacterEffect requires a TargetCharacterID and TargetLocation")
	}
	event := CharacterMovedEvent{
		CharacterID: payload.TargetCharacterID,
		NewLocation: payload.TargetLocation,
		Reason:      fmt.Sprintf("Ability: %s", self.Name),
	}
	return []Event{event}, nil
}

// AdjustParanoiaEffect 调整角色的妄想值。
type AdjustParanoiaEffect struct {
	Amount int `json:"amount"`
}

func (e *AdjustParanoiaEffect) ResolveChoices(ctx EffectContext, self *Ability) ([]TargetChoice, error) {
	return nil, nil
}

func (e *AdjustParanoiaEffect) Execute(ctx EffectContext, self *Ability, payload UseAbilityPayload) ([]Event, error) {
	if payload.TargetCharacterID == "" {
		return nil, fmt.Errorf("AdjustParanoiaEffect requires a TargetCharacterID")
	}
	event := ParanoiaAdjustedEvent{
		CharacterID: payload.TargetCharacterID,
		Amount:      e.Amount,
	}
	return []Event{event}, nil
}

// AdjustGoodwillEffect 调整角色的好感度。
type AdjustGoodwillEffect struct {
	Amount int `json:"amount"`
}

func (e *AdjustGoodwillEffect) ResolveChoices(ctx EffectContext, self *Ability) ([]TargetChoice, error) {
	return nil, nil
}

func (e *AdjustGoodwillEffect) Execute(ctx EffectContext, self *Ability, payload UseAbilityPayload) ([]Event, error) {
	if payload.TargetCharacterID == "" {
		return nil, fmt.Errorf("AdjustGoodgilEffect requires a TargetCharacterID")
	}
	event := GoodwillAdjustedEvent{
		CharacterID: payload.TargetCharacterID,
		Amount:      e.Amount,
	}
	return []Event{event}, nil
}

// AdjustIntrigueEffect 调整角色的阴谋值。
type AdjustIntrigueEffect struct {
	Amount int `json:"amount"`
}

func (e *AdjustIntrigueEffect) ResolveChoices(ctx EffectContext, self *Ability) ([]TargetChoice, error) {
	return nil, nil
}

func (e *AdjustIntrigueEffect) Execute(ctx EffectContext, self *Ability, payload UseAbilityPayload) ([]Event, error) {
	if payload.TargetCharacterID == "" {
		return nil, fmt.Errorf("AdjustIntrigueEffect requires a TargetCharacterID")
	}
	event := IntrigueAdjustedEvent{
		CharacterID: payload.TargetCharacterID,
		Amount:      e.Amount,
	}
	return []Event{event}, nil
}

// --- 自定义反序列化 ---

type effectWrapper struct {
	Type   EffectType      `json:"type"`
	Params json.RawMessage `json:"params"`
}

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
	default:
		return nil, fmt.Errorf("unknown effect type: '%s'", wrapper.Type)
	}

	if len(wrapper.Params) > 0 && string(wrapper.Params) != "null" {
		if err := json.Unmarshal(wrapper.Params, effect); err != nil {
			return nil, fmt.Errorf("failed to unmarshal params for type '%s': %w", wrapper.Type, err)
		}
	}

	return effect, nil
}
