package model

// TragedyType 定义可能发生的悲剧类型。
type TragedyType string

const (
	TragedyMurder  TragedyType = "Murder"
	TragedySuicide TragedyType = "Suicide"
	TragedySealed  TragedyType = "Sealed" // 例如：封印物品剧情
)

// TargetRuleType 定义悲剧如何选择目标角色。
type TargetRuleType string

const (
	TargetRuleSpecificCharacter      TargetRuleType = "SpecificCharacter"
	TargetRuleAnyCharacterAtLocation TargetRuleType = "AnyCharacterAtLocation"
)

// TragedyCondition 定义悲剧发生的条件。
type TragedyCondition struct {
	TragedyType TragedyType    `json:"tragedy_type"`
	Day         int            `json:"day"`         // 悲剧可能发生的日期
	CulpritID   string         `json:"culprit_id"`  // 导致此悲剧的嫌疑角色 ID
	Conditions  []Condition    `json:"conditions"`  // 必须满足的条件列表
	TargetRule  TargetRuleType `json:"target_rule"` // 悲剧如何选择目标角色
	IsActive    bool           `json:"-"`           // 运行时状态：此悲剧当前是否在剧本中活跃？
	IsPrevented bool           `json:"-"`           // 运行时状态：此悲剧是否已被阻止？
}

// Condition 定义悲剧的一个单一条件。
type Condition struct {
	CharacterID string       `json:"character_id"`
	Location    LocationType `json:"location"`
	MinParanoia int          `json:"min_paranoia"`
	IsAlone     bool         `json:"is_alone"` // 如果角色必须单独在某个地点，则为 true
	// 根据需要添加更多特定条件（例如，特定善意值，特定卡牌打出）
}
