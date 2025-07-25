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
	GameID              string                `json:"game_id"`                  // 游戏唯一标识符
	Script              Script                `json:"script"`                   // 当前使用的剧本
	Characters          map[string]*Character `json:"characters"`             // 游戏中所有角色的映射
	Players             map[string]*Player    `json:"players"`                // 游戏中所有玩家的映射
	CurrentDay          int                   `json:"current_day"`              // 当前天数
	CurrentLoop         int                   `json:"current_loop"`             // 当前循环次数
	CurrentPhase        GamePhase             `json:"current_phase"`            // 当前游戏阶段
	ActiveTragedies     map[TragedyType]bool  `json:"active_tragedies"`       // 此剧本中活跃的悲剧
	PreventedTragedies  map[TragedyType]bool  `json:"prevented_tragedies"`    // 此循环中已被阻止的悲剧
	PlayedCardsThisDay  map[string][]Card     `json:"played_cards_this_day"`  // 当天玩家打出的牌
	PlayedCardsThisLoop map[string][]Card     `json:"played_cards_this_loop"` // 本循环玩家打出的牌
	LastUpdateTime      time.Time             `json:"last_update_time"`        // 最后更新时间
	DayEvents           []GameEvent           `json:"day_events"`              // 当天发生的事件
	LoopEvents          []GameEvent           `json:"loop_events"`             // 本循环发生的事件
}

// GameEvent 表示游戏中发生的事件。
// 用于日志记录、广播和 LLM 上下文。
type GameEvent struct {
	Type      EventType   `json:"type"`      // 事件类型
	Payload   interface{} `json:"payload"`   // 事件负载
	Timestamp time.Time   `json:"timestamp"` // 事件时间戳
}
