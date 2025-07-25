package model

// TragedyType 定义可能发生的悲剧类型。
type TragedyType string

const (
	TragedyMurder  TragedyType = "Murder"  // 谋杀
	TragedySuicide TragedyType = "Suicide" // 自杀
	TragedySealed  TragedyType = "Sealed"  // 例如：封印物品剧情
)

// TargetRuleType 定义悲剧如何选择目标角色。
type TargetRuleType string

const (
	TargetRuleSpecificCharacter      TargetRuleType = "SpecificCharacter"      // 特定角色
	TargetRuleAnyCharacterAtLocation TargetRuleType = "AnyCharacterAtLocation" // 在某个地点的任何角色
)

// TragedyCondition 定义悲剧发生的条件。
type TragedyCondition struct {
	TragedyType TragedyType    `json:"tragedy_type"` // 悲剧类型
	Day         int            `json:"day"`          // 悲剧可能发生的日期
	CulpritID   string         `json:"culprit_id"`   // 嫌疑人ID
	Conditions  []Condition    `json:"conditions"`   // 必须满足的条件列表
	TargetRule  TargetRuleType `json:"target_rule"`  // 目标规则
	IsActive    bool           `json:"-"`            // 运行时状态：此悲剧当前是否在剧本中活跃？
	IsPrevented bool           `json:"-"`            // 运行时状态：此悲剧是否已被阻止？
}

// Condition 定义悲剧的一个单一条件。
type Condition struct {
	CharacterID string       `json:"character_id"` // 角色ID
	Location    LocationType `json:"location"`     // 所在位置
	MinParanoia int          `json:"min_paranoia"` // 最小妄想指数
	IsAlone     bool         `json:"is_alone"`     // 是否独处
}
