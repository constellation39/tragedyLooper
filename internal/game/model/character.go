package model

// RoleType 定义角色可能拥有的隐藏身份。
type RoleType string

const (
	RoleInnocent           RoleType = "Innocent"           // 无辜者
	RoleKiller             RoleType = "Killer"             // 杀手
	RoleBrain              RoleType = "Brain"              // 大脑
	RoleKeyPerson          RoleType = "KeyPerson"          // 关键人物
	RoleFriend             RoleType = "Friend"             // 朋友
	RoleConspiracyTheorist RoleType = "ConspiracyTheorist" // 阴谋论者
	RoleCultist            RoleType = "Cultist"            // 例如：具有善意拒绝的角色
	RoleMastermind         RoleType = "Mastermind"         // LLM 玩家角色
	RoleProtagonist        RoleType = "Protagonist"        // LLM 玩家角色
)

// Character 表示游戏中的一个角色。
type Character struct {
	ID              string       `json:"id"`               // 唯一标识符
	Name            string       `json:"name"`             // 角色名称
	Traits          []string     `json:"traits"`           // 角色特征，例如：["Student", "Journalist"]
	CurrentLocation LocationType `json:"current_location"` // 当前所在位置
	Paranoia        int          `json:"paranoia"`         // 妄想指数
	Goodwill        int          `json:"goodwill"`         // 好感度
	Intrigue        int          `json:"intrigue"`         // 阴谋标记
	IsAlive         bool         `json:"is_alive"`         // 是否存活
	Abilities       []Ability    `json:"abilities"`        // 角色能力
	HiddenRole      RoleType     `json:"-"`                // 角色的隐藏身份，对主角隐藏，仅供主谋查看
}

// CharacterConfig 定义特定剧本中角色的初始状态。
type CharacterConfig struct {
	CharacterID     string       `json:"character_id"`             // 引用基础角色定义
	InitialLocation LocationType `json:"initial_location"`         // 初始位置
	HiddenRole      RoleType     `json:"hidden_role"`              // 此剧本中角色的秘密身份
	IsCulpritFor    TragedyType  `json:"is_culprit_for,omitempty"` // 如果此角色是特定悲剧的嫌疑犯
}
