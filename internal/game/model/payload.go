package model

// PlayCardPayload 定义了打出卡牌动作所需的数据。
type PlayCardPayload struct {
	CardID            string       `json:"card_id"`
	TargetCharacterID string       `json:"target_character_id,omitempty"`
	TargetLocation    LocationType `json:"target_location,omitempty"`
}

// UseAbilityPayload 定义了使用能力动作所需的数据。
type UseAbilityPayload struct {
	CharacterID string `json:"character_id"`
	AbilityID   string `json:"ability_id"`
}

// MakeGuessPayload 定义了主角进行最终猜测时提交的数据结构。
type MakeGuessPayload struct {
	// GuessedRoles 是一个映射，键是角色ID，值是猜测的角色身份 (e.g., "KeyPerson", "Killer")。
	GuessedRoles map[string]RoleType `json:"guessed_roles"`
}
