package model

import (
	"encoding/json"
	"fmt"
)

// EffectType 定义能力或卡牌可能产生的效果类型。
// 这是组合方法的核心。
type EffectType string

const (
	EffectTypeMoveCharacter        EffectType = "MoveCharacter"        // 移动角色
	EffectTypeAdjustParanoia       EffectType = "AdjustParanoia"       // 调整妄想
	EffectTypeAdjustGoodwill       EffectType = "AdjustGoodwill"       // 调整好感
	EffectTypeAdjustIntrigue       EffectType = "AdjustIntrigue"       // 调整阴谋
	EffectTypeRevealRole           EffectType = "RevealRole"           // 用于特定能力，揭示角色
	EffectTypePreventTragedy       EffectType = "PreventTragedy"       // 阻止悲剧
	EffectTypeGrantAbility         EffectType = "GrantAbility"         // 授予能力
	EffectTypeCheckLocationAlone   EffectType = "CheckLocationAlone"   // 用于悲剧条件，检查地点是否有人独处
	EffectTypeCheckCharacterStatus EffectType = "CheckCharacterStatus" // 用于悲剧条件，检查角色状态
)

// AbilityTriggerType 定义能力何时可以被触发。
type AbilityTriggerType string

const (
	AbilityTriggerDayStart        AbilityTriggerType = "DayStart"        // 天开始时
	AbilityTriggerMastermindPhase AbilityTriggerType = "MastermindPhase" // 主谋阶段
	AbilityTriggerGoodwillPhase   AbilityTriggerType = "GoodwillPhase"   // 好感阶段
	AbilityTriggerPassive         AbilityTriggerType = "Passive"         // 被动
)

const (
	EventChoiceRequired EventType = "ChoiceRequired" // 需要玩家进行选择
)

// Ability 定义角色的特殊技能。
type Ability struct {
	Name         string             `json:"name"`                   // 能力名称
	Description  string             `json:"description"`            // 能力描述
	TriggerType  AbilityTriggerType `json:"trigger_type"`           // 触发时机
	Effect       Effect             `json:"effect"`                 // 实际效果
	OncePerLoop  bool               `json:"once_per_loop"`          // 每循环只能使用一次
	RefusalRole  RoleType           `json:"refusal_role,omitempty"` // 如果有，指定拒绝此善意能力的特定角色身份
	UsedThisLoop bool               `json:"-"`                      // 运行时状态，不用于配置
}

// UnmarshalJSON for Ability to handle the polymorphic Effect interface.
func (a *Ability) UnmarshalJSON(data []byte) error {
	type Alias Ability
	aux := &struct {
		Effect json.RawMessage `json:"effect"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal ability shell: %w", err)
	}

	effect, err := UnmarshalEffect(aux.Effect)
	if err != nil {
		return fmt.Errorf("failed to unmarshal effect for ability '%s': %w", a.Name, err)
	}
	a.Effect = effect

	return nil
}

// MarshalJSON for Ability to handle the polymorphic Effect interface.
func (a *Ability) MarshalJSON() ([]byte, error) {
	type Alias Ability
	aux := &struct {
		Effect interface{} `json:"effect"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	// Find the EffectType for the concrete Effect implementation
	var effectType EffectType
	// This is a bit brittle, but it's a common way to handle this in Go.
	// A better way might be to have a Type() method on the Effect interface.
	switch a.Effect.(type) {
	case *MoveCharacterEffect:
		effectType = EffectTypeMoveCharacter
	case *AdjustParanoiaEffect:
		effectType = EffectTypeAdjustParanoia
	case *AdjustGoodwillEffect:
		effectType = EffectTypeAdjustGoodwill
	case *AdjustIntrigueEffect:
		effectType = EffectTypeAdjustIntrigue
	default:
		return nil, fmt.Errorf("unknown effect type for marshaling")
	}

	// Wrap the effect in a struct with its type
	aux.Effect = struct {
		Type   EffectType  `json:"type"`
		Params interface{} `json:"params"`
	}{
		Type:   effectType,
		Params: a.Effect,
	}

	return json.Marshal(aux)
}
