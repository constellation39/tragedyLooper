package model

// LocationType 定义游戏中的可能地点。
type LocationType string

const (
	LocationHospital LocationType = "Hospital"
	LocationShrine   LocationType = "Shrine"
	LocationCity     LocationType = "City"
	LocationSchool   LocationType = "School"
)

// RoleType 定义角色可能拥有的隐藏身份。
type RoleType string

const (
	RoleInnocent           RoleType = "Innocent"
	RoleKiller             RoleType = "Killer"
	RoleBrain              RoleType = "Brain"
	RoleKeyPerson          RoleType = "KeyPerson"
	RoleFriend             RoleType = "Friend"
	RoleConspiracyTheorist RoleType = "ConspiracyTheorist"
	RoleCultist            RoleType = "Cultist"     // 例如：具有善意拒绝的角色
	RoleMastermind         RoleType = "Mastermind"  // LLM 玩家角色
	RoleProtagonist        RoleType = "Protagonist" // LLM 玩家角色
)

// CardType 定义玩家可以打出的卡牌类型。
type CardType string

const (
	CardTypeMovement CardType = "Movement"
	CardTypeParanoia CardType = "Paranoia"
	CardTypeGoodwill CardType = "Goodwill"
	CardTypeIntrigue CardType = "Intrigue"
	CardTypeSpecial  CardType = "Special" // 用于特殊卡牌效果
)

// EffectType 定义能力或卡牌可能产生的效果类型。
// 这是组合方法的核心。
type EffectType string

const (
	EffectTypeMoveCharacter        EffectType = "MoveCharacter"
	EffectTypeAdjustParanoia       EffectType = "AdjustParanoia"
	EffectTypeAdjustGoodwill       EffectType = "AdjustGoodwill"
	EffectTypeAdjustIntrigue       EffectType = "AdjustIntrigue"
	EffectTypeRevealRole           EffectType = "RevealRole" // 用于特定能力
	EffectTypePreventTragedy       EffectType = "PreventTragedy"
	EffectTypeGrantAbility         EffectType = "GrantAbility"
	EffectTypeCheckLocationAlone   EffectType = "CheckLocationAlone"   // 用于悲剧条件
	EffectTypeCheckCharacterStatus EffectType = "CheckCharacterStatus" // 用于悲剧条件
)

// AbilityTriggerType 定义能力何时可以被触发。
type AbilityTriggerType string

const (
	AbilityTriggerDayStart        AbilityTriggerType = "DayStart"
	AbilityTriggerMastermindPhase AbilityTriggerType = "MastermindPhase"
	AbilityTriggerGoodwillPhase   AbilityTriggerType = "GoodwillPhase"
	AbilityTriggerPassive         AbilityTriggerType = "Passive"
)

// TragedyType 定义可能发生的悲剧类型。
type TragedyType string

const (
	TragedyMurder  TragedyType = "Murder"
	TragedySuicide TragedyType = "Suicide"
	TragedySealed  TragedyType = "Sealed" // 例如：封印物品剧情
)

// TargetRuleType 定义悲剧如何选择目标角色。
type TargetRuleType string

const (
	TargetRuleSpecificCharacter      TargetRuleType = "SpecificCharacter"
	TargetRuleAnyCharacterAtLocation TargetRuleType = "AnyCharacterAtLocation"
)

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
