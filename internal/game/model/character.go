package model

// RoleType 定义角色可能拥有的隐藏身份。
type RoleType string

const (
	RoleInnocent           RoleType = "Innocent"
	RoleKiller             RoleType = "Killer"
	RoleBrain              RoleType = "Brain"
	RoleKeyPerson          RoleType = "KeyPerson"
	RoleFriend             RoleType = "Friend"
	RoleConspiracyTheorist RoleType = "ConspiracyTheorist"
	RoleCultist            RoleType = "Cultist"     // 例如：具有善意拒绝的角色
	RoleMastermind         RoleType = "Mastermind"  // LLM 玩家角色
	RoleProtagonist        RoleType = "Protagonist" // LLM 玩家角色
)

// Character 表示游戏中的一个角色。
type Character struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Traits          []string     `json:"traits"` // 例如：["Student", "Journalist"]
	CurrentLocation LocationType `json:"current_location"`
	Paranoia        int          `json:"paranoia"`
	Goodwill        int          `json:"goodwill"`
	Intrigue        int          `json:"intrigue"` // 主谋的阴谋标记
	IsAlive         bool         `json:"is_alive"`
	Abilities       []Ability    `json:"abilities"`
	HiddenRole      RoleType     `json:"-"` // 对主角隐藏，仅供主谋查看
}

// CharacterConfig 定义特定剧本中角色的初始状态。
type CharacterConfig struct {
	CharacterID     string       `json:"character_id"` // 引用基础角色定义
	InitialLocation LocationType `json:"initial_location"`
	HiddenRole      RoleType     `json:"hidden_role"`              // 此剧本中角色的秘密身份
	IsCulpritFor    TragedyType  `json:"is_culprit_for,omitempty"` // 如果此角色是特定悲剧的嫌疑犯
}
