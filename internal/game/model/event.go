package model

// Event 代表游戏中发生的一个原子性、不可变的事实。
type Event interface {
	IsEvent()
}

// --- 具体事件定义 ---

// CharacterMovedEvent 记录了一个角色被移动到新地点的事实。
type CharacterMovedEvent struct {
	CharacterID string       `json:"character_id"`
	NewLocation LocationType `json:"new_location"`
	Reason      string       `json:"reason,omitempty"`
}

func (e CharacterMovedEvent) IsEvent() {}

// ParanoiaAdjustedEvent 记录了角色妄想值发生变化的事实。
type ParanoiaAdjustedEvent struct {
	CharacterID string `json:"character_id"`
	Amount      int    `json:"amount"`
}

func (e ParanoiaAdjustedEvent) IsEvent() {}

// GoodwillAdjustedEvent 记录了角色好感度发生变化的事实。
type GoodwillAdjustedEvent struct {
	CharacterID string `json:"character_id"`
	Amount      int    `json:"amount"`
}

func (e GoodwillAdjustedEvent) IsEvent() {}

// IntrigueAdjustedEvent 记录了角色阴谋值发生变化的事实。
type IntrigueAdjustedEvent struct {
	CharacterID string `json:"character_id"`
	Amount      int    `json:"amount"`
}

func (e IntrigueAdjustedEvent) IsEvent() {}
