package model

// EffectType 定义能力或卡牌可能产生的效果类型。
// 这是组合方法的核心。
type EffectType string

const (
	EffectTypeMoveCharacter        EffectType = "MoveCharacter"
	EffectTypeAdjustParanoia       EffectType = "AdjustParanoia"
	EffectTypeAdjustGoodwill       EffectType = "AdjustGoodwill"
	EffectTypeAdjustIntrigue       EffectType = "AdjustIntrigue"
	EffectTypeRevealRole           EffectType = "RevealRole" // 用于特定能力
	EffectTypePreventTragedy       EffectType = "PreventTragedy"
	EffectTypeGrantAbility         EffectType = "GrantAbility"
	EffectTypeCheckLocationAlone   EffectType = "CheckLocationAlone"   // 用于悲剧条件
	EffectTypeCheckCharacterStatus EffectType = "CheckCharacterStatus" // 用于悲剧条件
)

// AbilityTriggerType 定义能力何时可以被触发。
type AbilityTriggerType string

const (
	AbilityTriggerDayStart        AbilityTriggerType = "DayStart"
	AbilityTriggerMastermindPhase AbilityTriggerType = "MastermindPhase"
	AbilityTriggerGoodwillPhase   AbilityTriggerType = "GoodwillPhase"
	AbilityTriggerPassive         AbilityTriggerType = "Passive"
)

// AbilityEffect 定义能力或卡牌的具体效果。
// 这使用类型和参数的组合方法。
type AbilityEffect struct {
	Type   EffectType             `json:"type"`   // 例如："MoveCharacter", "AdjustParanoia"
	Params map[string]interface{} `json:"params"` // 例如：{"location": "School", "amount": 1}
}

// Ability 定义角色的特殊技能。
type Ability struct {
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	TriggerType  AbilityTriggerType `json:"trigger_type"`           // 何时可以使用/触发
	Effect       AbilityEffect      `json:"effect"`                 // 实际效果
	OncePerLoop  bool               `json:"once_per_loop"`          // 每循环只能使用一次
	RefusalRole  RoleType           `json:"refusal_role,omitempty"` // 如果有，指定拒绝此善意能力的特定角色身份
	UsedThisLoop bool               `json:"-"`                      // 运行时状态，不用于配置
}

// UseAbilityPayload for ActionUseAbility
type UseAbilityPayload struct {
	AbilityName       string       `json:"ability_name"`
	TargetCharacterID string       `json:"target_character_id,omitempty"`
	TargetLocation    LocationType `json:"target_location,omitempty"`
}
