package model

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

// AbilityEffect 定义能力或卡牌的具体效果。
// 这使用类型和参数的组合方法。
type AbilityEffect struct {
	Type   EffectType             `json:"type"`   // 效果类型，例如："MoveCharacter", "AdjustParanoia"
	Params map[string]interface{} `json:"params"` // 效果参数，例如：{"location": "School", "amount": 1}
}

// Ability 定义角色的特殊技能。
type Ability struct {
	Name         string             `json:"name"`                   // 能力名称
	Description  string             `json:"description"`            // 能力描述
	TriggerType  AbilityTriggerType `json:"trigger_type"`           // 触发时机
	Effect       AbilityEffect      `json:"effect"`                 // 实际效果
	OncePerLoop  bool               `json:"once_per_loop"`          // 每循环只能使用一次
	RefusalRole  RoleType           `json:"refusal_role,omitempty"` // 如果有，指定拒绝此善意能力的特定角色身份
	UsedThisLoop bool               `json:"-"`                      // 运行时状态，不用于配置
}

// UseAbilityPayload for ActionUseAbility
type UseAbilityPayload struct {
	AbilityName       string       `json:"ability_name"`                 // 使用的能力名称
	TargetCharacterID string       `json:"target_character_id,omitempty"` // 目标角色ID
	TargetLocation    LocationType `json:"target_location,omitempty"`    // 目标位置
}