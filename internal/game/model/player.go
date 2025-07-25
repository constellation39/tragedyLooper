package model

// PlayerActionType 定义玩家可以采取的行动类型。
type PlayerActionType string

const (
	ActionPlayCard          PlayerActionType = "PlayCard"
	ActionUseAbility        PlayerActionType = "UseAbility"
	ActionMakeGuess         PlayerActionType = "MakeGuess" // 主角最终猜测
	ActionEndTurn           PlayerActionType = "EndTurn"
	ActionReadyForNextPhase PlayerActionType = "ReadyForNextPhase"
)

// PlayerRole 定义连接玩家的角色（人类或 LLM）。
type PlayerRole string

const (
	PlayerRoleMastermind  PlayerRole = "Mastermind"
	PlayerRoleProtagonist PlayerRole = "Protagonist"
)

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

// MakeGuessPayload for ActionMakeGuess
type MakeGuessPayload struct {
	GuessedRoles map[string]RoleType `json:"guessed_roles"` // CharacterID -> GuessedRole
}
