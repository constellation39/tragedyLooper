package model

import "time"

// GamePhase 定义游戏日期的当前阶段。
type GamePhase string

const (
	PhaseMorning          GamePhase = "Morning"          // 早晨
	PhaseCardPlay         GamePhase = "CardPlay"         // 出牌阶段
	PhaseCardReveal       GamePhase = "CardReveal"       // 牌揭示阶段
	PhaseCardResolve      GamePhase = "CardResolve"      // 牌结算阶段
	PhaseAbilities        GamePhase = "Abilities"        // 能力阶段
	PhaseIncidents        GamePhase = "Incidents"        // 事件阶段
	PhaseDayEnd           GamePhase = "DayEnd"           // 天结束
	PhaseLoopEnd          GamePhase = "LoopEnd"          // 循环结束
	PhaseGameOver         GamePhase = "GameOver"         // 游戏结束
	PhaseProtagonistGuess GamePhase = "ProtagonistGuess" // 最终猜测阶段
)

// EventType for GameEvent
type EventType string

const (
	EventCardPlayed       EventType = "CardPlayed"       // 打出卡牌
	EventCharacterMoved   EventType = "CharacterMoved"   // 角色移动
	EventParanoiaAdjusted EventType = "ParanoiaAdjusted" // 妄想调整
	EventGoodwillAdjusted EventType = "GoodwillAdjusted" // 好感调整
	EventIntrigueAdjusted EventType = "IntrigueAdjusted" // 阴谋调整
	EventAbilityUsed      EventType = "AbilityUsed"      // 使用能力
	EventTragedyTriggered EventType = "TragedyTriggered" // 悲剧触发
	EventTragedyPrevented EventType = "TragedyPrevented" // 悲剧阻止
	EventDayAdvanced      EventType = "DayAdvanced"      // 天数推进
	EventLoopReset        EventType = "LoopReset"        // 循环重置
	EventGameOver         EventType = "GameOver"         // 游戏结束
	EventPlayerGuess      EventType = "PlayerGuess"      // 玩家猜测
)

// GameState 表示游戏实例的权威当前状态。
// 这包含所有信息，包括主谋的隐藏信息。
type GameState struct {
	GameID              string                `json:"game_id"`                // 游戏唯一标识符
	Script              Script                `json:"script"`                 // 当前使用的剧本
	Characters          map[string]*Character `json:"characters"`             // 游戏中所有角色的映射
	Players             map[string]*Player    `json:"players"`                // 游戏中所有玩家的映射
	CurrentDay          int                   `json:"current_day"`            // 当前天数
	CurrentLoop         int                   `json:"current_loop"`           // 当前循环次数
	CurrentPhase        GamePhase             `json:"current_phase"`          // 当前游戏阶段
	ActiveTragedies     map[TragedyType]bool  `json:"active_tragedies"`       // 此剧本中活跃的悲剧
	PreventedTragedies  map[TragedyType]bool  `json:"prevented_tragedies"`    // 此循环中已被阻止的悲剧
	PlayedCardsThisDay  map[string][]Card     `json:"played_cards_this_day"`  // 当天玩家打出的牌
	PlayedCardsThisLoop map[string][]Card     `json:"played_cards_this_loop"` // 本循环玩家打出的牌
	LastUpdateTime      time.Time             `json:"last_update_time"`       // 最后更新时间
	DayEvents           []GameEvent           `json:"day_events"`             // 当天发生的事件
	LoopEvents          []GameEvent           `json:"loop_events"`            // 本循环发生的事件
}

// GameEvent 表示游戏中发生的事件。
// 用于日志记录、广播和 LLM 上下文。
type GameEvent struct {
	Type      EventType   `json:"type"`      // 事件类型
	Payload   interface{} `json:"payload"`   // 事件负载
	Timestamp time.Time   `json:"timestamp"` // 事件时间戳
}

// ActionType 定义了玩家可以执行的操作类型。
type ActionType string

const (
	ActionPlayCard          ActionType = "PlayCard"          // 出牌
	ActionUseAbility        ActionType = "UseAbility"        // 使用能力
	ActionMakeGuess         ActionType = "MakeGuess"         // 作出猜测
	ActionReadyForNextPhase ActionType = "ReadyForNextPhase" // 准备好进入下一阶段
)

// PlayerAction 代表单个玩家提交的动作。
type PlayerAction struct {
	PlayerID string      `json:"player_id"`
	GameID   string      `json:"game_id"`
	Type     ActionType  `json:"type"`
	Payload  interface{} `json:"payload"` // e.g., PlayCardPayload, UseAbilityPayload
}

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
