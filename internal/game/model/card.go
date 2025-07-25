package model

// CardType 定义玩家可以打出的卡牌类型。
type CardType string

const (
	CardTypeMovement CardType = "Movement" // 移动
	CardTypeParanoia CardType = "Paranoia" // 妄想
	CardTypeGoodwill CardType = "Goodwill" // 好感
	CardTypeIntrigue CardType = "Intrigue" // 阴谋
	CardTypeSpecial  CardType = "Special"  // 用于特殊卡牌效果
)

// Card 表示一张行动卡。
type Card struct {
	ID                string        `json:"id"`                             // 唯一标识符
	Name              string        `json:"name"`                           // 卡牌名称
	CardType          CardType      `json:"card_type"`                       // 卡牌类型
	OwnerRole         PlayerRole    `json:"owner_role"`                     // 所属玩家角色（主谋或主角）
	TargetType        string        `json:"target_type"`                    // 目标类型（"Character" 或 "Location"）
	Effect            AbilityEffect `json:"effect"`                         // 卡牌效果
	OncePerLoop       bool          `json:"once_per_loop"`                  // 每循环只能使用一次
	UsedThisLoop      bool          `json:"-"`                              // 运行时状态
	TargetCharacterID string        `json:"target_character_id,omitempty"`  // 目标角色ID
	TargetLocation    LocationType  `json:"target_location,omitempty"`     // 目标位置
}

// PlayCardPayload for ActionPlayCard
type PlayCardPayload struct {
	CardID            string       `json:"card_id"`                        // 打出的卡牌ID
	TargetCharacterID string       `json:"target_character_id,omitempty"` // 目标角色ID
	TargetLocation    LocationType `json:"target_location,omitempty"`    // 目标位置
}