package model

// PlayerActionType 定义玩家可以采取的行动类型。
type PlayerActionType string

const (
	ActionPlayCard          PlayerActionType = "PlayCard"          // 出牌
	ActionUseAbility        PlayerActionType = "UseAbility"        // 使用能力
	ActionMakeGuess         PlayerActionType = "MakeGuess"         // 主角最终猜测
	ActionEndTurn           PlayerActionType = "EndTurn"           // 结束回合
	ActionReadyForNextPhase PlayerActionType = "ReadyForNextPhase" // 准备好进入下一阶段
)

// PlayerRole 定义连接玩家的角色（人类或 LLM）。
type PlayerRole string

const (
	PlayerRoleMastermind  PlayerRole = "Mastermind"  // 主谋
	PlayerRoleProtagonist PlayerRole = "Protagonist" // 主角
)

// Player 表示一个连接的玩家（人类或 LLM）。
type Player struct {
	ID                 string                 `json:"id"`                       // 唯一标识符
	Name               string                 `json:"name"`                     // 玩家名称
	Role               PlayerRole             `json:"role"`                     // 玩家角色（主谋或主角）
	IsLLM              bool                   `json:"is_llm"`                   // 是否为LLM玩家
	Hand               []Card                 `json:"hand"`                     // 玩家手牌
	DeductionKnowledge map[string]interface{} `json:"deduction_knowledge"`      // 主角的推理知识
	LLMSessionID       string                 `json:"llm_session_id,omitempty"` // LLM玩家的会话ID
}

// PlayerView 表示特定玩家的游戏状态过滤视图。
// 这是发送给客户端（人类或 LLM）的内容。
type PlayerView struct {
	GameID             string                 `json:"game_id"`                   // 游戏唯一标识符
	ScriptID           string                 `json:"script_id"`                 // 剧本ID
	Characters         map[string]*Character  `json:"characters"`                // 角色信息（隐藏身份已移除）
	Players            map[string]*Player     `json:"players"`                   // 玩家信息（敏感信息已移除）
	CurrentDay         int                    `json:"current_day"`               // 当前天数
	CurrentLoop        int                    `json:"current_loop"`              // 当前循环次数
	CurrentPhase       GamePhase              `json:"current_phase"`             // 当前游戏阶段
	ActiveTragedies    map[TragedyType]bool   `json:"active_tragedies"`          // 活跃的悲剧（公开信息）
	PreventedTragedies map[TragedyType]bool   `json:"prevented_tragedies"`       // 已阻止的悲剧（公开信息）
	YourHand           []Card                 `json:"your_hand,omitempty"`       // 你的手牌（仅对请求玩家可见）
	YourDeductions     map[string]interface{} `json:"your_deductions,omitempty"` // 你的推理（仅对主角可见）
	PublicEvents       []GameEvent            `json:"public_events"`             // 公开事件
}

// PlayerAction 表示从玩家客户端发送到服务器的动作。
type PlayerAction struct {
	PlayerID string           `json:"player_id"` // 玩家ID
	GameID   string           `json:"game_id"`   // 游戏ID
	Type     PlayerActionType `json:"type"`      // 动作类型
	Payload  interface{}      `json:"payload"`   // 动作负载
}

// MakeGuessPayload for ActionMakeGuess
type MakeGuessPayload struct {
	GuessedRoles map[string]RoleType `json:"guessed_roles"` // 猜测的角色身份映射 (CharacterID -> GuessedRole)
}
