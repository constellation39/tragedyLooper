package model

import "time"

// GamePhase 定义游戏日期的当前阶段。
type GamePhase string

const (
	PhaseMorning          GamePhase = "Morning"
	PhaseCardPlay         GamePhase = "CardPlay"
	PhaseCardReveal       GamePhase = "CardReveal"
	PhaseCardResolve      GamePhase = "CardResolve"
	PhaseAbilities        GamePhase = "Abilities"
	PhaseIncidents        GamePhase = "Incidents"
	PhaseDayEnd           GamePhase = "DayEnd"
	PhaseLoopEnd          GamePhase = "LoopEnd"
	PhaseGameOver         GamePhase = "GameOver"
	PhaseProtagonistGuess GamePhase = "ProtagonistGuess" // 最终猜测阶段
)

// EventType for GameEvent
type EventType string

const (
	EventCardPlayed       EventType = "CardPlayed"
	EventCharacterMoved   EventType = "CharacterMoved"
	EventParanoiaAdjusted EventType = "ParanoiaAdjusted"
	EventGoodwillAdjusted EventType = "GoodwillAdjusted"
	EventIntrigueAdjusted EventType = "IntrigueAdjusted"
	EventAbilityUsed      EventType = "AbilityUsed"
	EventTragedyTriggered EventType = "TragedyTriggered"
	EventTragedyPrevented EventType = "TragedyPrevented"
	EventDayAdvanced      EventType = "DayAdvanced"
	EventLoopReset        EventType = "LoopReset"
	EventGameOver         EventType = "GameOver"
	EventPlayerGuess      EventType = "PlayerGuess"
)

// GameState 表示游戏实例的权威当前状态。
// 这包含所有信息，包括主谋的隐藏信息。
type GameState struct {
	GameID              string                `json:"game_id"`
	Script              Script                `json:"script"`
	Characters          map[string]*Character `json:"characters"` // 角色 ID 到 Character 对象的映射
	Players             map[string]*Player    `json:"players"`    // 玩家 ID 到 Player 对象的映射
	CurrentDay          int                   `json:"current_day"`
	CurrentLoop         int                   `json:"current_loop"`
	CurrentPhase        GamePhase             `json:"current_phase"`
	ActiveTragedies     map[TragedyType]bool  `json:"active_tragedies"`       // 此剧本中活跃的悲剧
	PreventedTragedies  map[TragedyType]bool  `json:"prevented_tragedies"`    // 此循环中已被阻止的悲剧
	PlayedCardsThisDay  map[string][]Card     `json:"played_cards_this_day"`  // PlayerID -> 打出的牌
	PlayedCardsThisLoop map[string][]Card     `json:"played_cards_this_loop"` // PlayerID -> 打出的牌（用于每循环一次跟踪）
	// 添加其他全局游戏状态变量
	LastUpdateTime time.Time `json:"last_update_time"`
	// 此天/循环中发生的事件，用于 LLM 上下文和人类回顾
	DayEvents  []GameEvent `json:"day_events"`
	LoopEvents []GameEvent `json:"loop_events"`
}

// GameEvent 表示游戏中发生的事件。
// 用于日志记录、广播和 LLM 上下文。
type GameEvent struct {
	Type      EventType   `json:"type"`    // 例如："CardPlayed", "CharacterMoved", "TragedyTriggered"
	Payload   interface{} `json:"payload"` // 事件的具体数据
	Timestamp time.Time   `json:"timestamp"`
	// 添加字段以指示可见性（例如，"public", "mastermind_only"）
}
