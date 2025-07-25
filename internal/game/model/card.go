package model

// CardType 定义玩家可以打出的卡牌类型。
type CardType string

const (
	CardTypeMovement CardType = "Movement"
	CardTypeParanoia CardType = "Paranoia"
	CardTypeGoodwill CardType = "Goodwill"
	CardTypeIntrigue CardType = "Intrigue"
	CardTypeSpecial  CardType = "Special" // 用于特殊卡牌效果
)

// Card 表示一张行动卡。
type Card struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	CardType          CardType      `json:"card_type"`
	OwnerRole         PlayerRole    `json:"owner_role"`    // 主谋或主角卡
	TargetType        string        `json:"target_type"`   // "Character" 或 "Location"
	Effect            AbilityEffect `json:"effect"`        // 卡牌效果，结构与 AbilityEffect 相同
	OncePerLoop       bool          `json:"once_per_loop"` // 每循环只能使用一次
	UsedThisLoop      bool          `json:"-"`             // 运行时状态
	TargetCharacterID string        `json:"target_character_id,omitempty"`
	TargetLocation    LocationType  `json:"target_location,omitempty"`
}

// PlayCardPayload for ActionPlayCard
type PlayCardPayload struct {
	CardID            string       `json:"card_id"`
	TargetCharacterID string       `json:"target_character_id,omitempty"`
	TargetLocation    LocationType `json:"target_location,omitempty"`
}
