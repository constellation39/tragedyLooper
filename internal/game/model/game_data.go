package model

import "time"

// AbilityEffect 定义能力或卡牌的具体效果。
// 这使用类型和参数的组合方法。
type AbilityEffect struct {
	Type   EffectType             `json:"type"`   // 例如："MoveCharacter", "AdjustParanoia"
	Params map[string]interface{} `json:"params"` // 例如：{"location": "School", "amount": 1}
}

// Ability 定义角色的特殊技能。
type Ability struct {
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	TriggerType  AbilityTriggerType `json:"trigger_type"`           // 何时可以使用/触发
	Effect       AbilityEffect      `json:"effect"`                 // 实际效果
	OncePerLoop  bool               `json:"once_per_loop"`          // 每循环只能使用一次
	RefusalRole  RoleType           `json:"refusal_role,omitempty"` // 如果有，指定拒绝此善意能力的特定角色身份
	UsedThisLoop bool               `json:"-"`                      // 运行时状态，不用于配置
}

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

// Script 定义一个特定的游戏场景。
type Script struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	MainPlot    string             `json:"main_plot"`     // 例如："谋杀计划", "封印物品"
	SubPlots    []string           `json:"sub_plots"`     // 例如："朋友圈", "阴谋"
	Characters  []CharacterConfig  `json:"characters"`    // 此剧本的初始角色配置
	Tragedies   []TragedyCondition `json:"tragedies"`     // 此剧本预定义的悲剧
	LoopCount   int                `json:"loop_count"`    // 允许的总循环次数
	DaysPerLoop int                `json:"days_per_loop"` // 每循环天数
	// 添加任何剧本特定的规则或初始设置
}

// CharacterConfig 定义特定剧本中角色的初始状态。
type CharacterConfig struct {
	CharacterID     string       `json:"character_id"` // 引用基础角色定义
	InitialLocation LocationType `json:"initial_location"`
	HiddenRole      RoleType     `json:"hidden_role"`              // 此剧本中角色的秘密身份
	IsCulpritFor    TragedyType  `json:"is_culprit_for,omitempty"` // 如果此角色是特定悲剧的嫌疑犯
}

// Player 表示一个连接的玩家（人类或 LLM）。
type Player struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Role  PlayerRole `json:"role"` // 主谋或主角
	IsLLM bool       `json:"is_llm"`
	Hand  []Card     `json:"hand"` // 玩家手中的牌
	// 对于主角，这将存储他们在循环中积累的推理知识。
	// 这是在循环中持续存在的“推理知识”。
	DeductionKnowledge map[string]interface{} `json:"deduction_knowledge"`
	// 对于 LLM 玩家，这可能还包括会话 ID 或特定的 LLM 配置。
	LLMSessionID string `json:"llm_session_id,omitempty"`
}

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

// PlayerView 表示特定玩家的游戏状态过滤视图。
// 这是发送给客户端（人类或 LLM）的内容。
type PlayerView struct {
	GameID             string                 `json:"game_id"`
	ScriptID           string                 `json:"script_id"`
	Characters         map[string]*Character  `json:"characters"` // 角色（隐藏身份已移除）
	Players            map[string]*Player     `json:"players"`    // 玩家（敏感信息已移除，例如其他玩家的手牌）
	CurrentDay         int                    `json:"current_day"`
	CurrentLoop        int                    `json:"current_loop"`
	CurrentPhase       GamePhase              `json:"current_phase"`
	ActiveTragedies    map[TragedyType]bool   `json:"active_tragedies"`          // 仅公开信息
	PreventedTragedies map[TragedyType]bool   `json:"prevented_tragedies"`       // 仅公开信息
	YourHand           []Card                 `json:"your_hand,omitempty"`       // 仅对请求玩家可见
	YourDeductions     map[string]interface{} `json:"your_deductions,omitempty"` // 仅对主角可见
	PublicEvents       []GameEvent            `json:"public_events"`             // 对所有人可见的事件
	// 添加其他公开游戏状态变量
}

// PlayerAction 表示从玩家客户端发送到服务器的动作。
type PlayerAction struct {
	PlayerID string           `json:"player_id"`
	GameID   string           `json:"game_id"`
	Type     PlayerActionType `json:"type"`    // 例如："PlayCard", "UseAbility"
	Payload  interface{}      `json:"payload"` // 动作的具体数据（例如，CardID, TargetCharacterID）
}

// PlayCardPayload for ActionPlayCard
type PlayCardPayload struct {
	CardID            string       `json:"card_id"`
	TargetCharacterID string       `json:"target_character_id,omitempty"`
	TargetLocation    LocationType `json:"target_location,omitempty"`
}

// UseAbilityPayload for ActionUseAbility
type UseAbilityPayload struct {
	AbilityName       string       `json:"ability_name"`
	TargetCharacterID string       `json:"target_character_id,omitempty"`
	TargetLocation    LocationType `json:"target_location,omitempty"`
}

// MakeGuessPayload for ActionMakeGuess
type MakeGuessPayload struct {
	GuessedRoles map[string]RoleType `json:"guessed_roles"` // CharacterID -> GuessedRole
}
