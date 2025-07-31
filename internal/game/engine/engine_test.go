package engine

import (
	"context"
	"testing"
	"time"
	"tragedylooper/internal/game/loader"
	"tragedylooper/internal/logger"
	v1 "tragedylooper/pkg/proto/tragedylooper/v1"

	"github.com/stretchr/testify/assert"
)

// mockActionGenerator 是一个用于测试的简单 AI 行动生成器。
type mockActionGenerator struct{}

func (m *mockActionGenerator) GenerateAction(ctx context.Context, agc *ActionGeneratorContext) (*v1.PlayerActionPayload, error) {
	// 对于此测试，我们不需要复杂的 AI 逻辑。
	// 我们可以返回一个简单的跳过行动。
	return &v1.PlayerActionPayload{
		Payload: &v1.PlayerActionPayload_PassTurn{
			PassTurn: &v1.PassTurnAction{},
		},
	}, nil
}

func helper_NewGameEngineForTest(t *testing.T) *GameEngine {
	t.Helper()

	log := logger.New()

	gameConfig, err := loader.LoadConfig("../../../data", "first_steps")
	if err != nil {
		t.Fatalf("failed to load game data: %v", err)
	}

	players := []*v1.Player{
		{Id: 1, Name: "Mastermind", Role: v1.PlayerRole_PLAYER_ROLE_MASTERMIND, IsLlm: true},
		{Id: 2, Name: "Protagonist 1", Role: v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, IsLlm: true},
		{Id: 3, Name: "Protagonist 2", Role: v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, IsLlm: true},
	}

	engine, err := NewGameEngine(log, players, &mockActionGenerator{}, gameConfig)
	if err != nil {
		t.Fatalf("failed to create game engine: %v", err)
	}

	return engine
}

// TestEngine_Integration_CardPlayAndIncidentTrigger 验证整个数据流：
// 1. 从 JSON 文件加载游戏数据。
// 2. 初始化引擎。
// 3. 玩家打出一张牌以增加角色的偏执。
// 4. 验证偏执状态已更新。
// 5. 继续增加偏执，直到触发事件。
// 6. 验证事件的效果已应用（角色获得新特征）。
func TestEngine_Integration_CardPlayAndIncidentTrigger(t *testing.T) {
	engine := helper_NewGameEngineForTest(t)

	// --- 设置：预先放置角色并准备主谋的手牌 ---
	mastermind := engine.getPlayerByID(1)
	doctor := engine.GetCharacterByID(2)
	highSchoolGirl := engine.GetCharacterByID(1)
	assert.NotNil(t, mastermind)
	assert.NotNil(t, doctor)
	assert.NotNil(t, highSchoolGirl)

	// 将高中女生移动到与医生相同的位置（医院）
	doctor.CurrentLocation = highSchoolGirl.CurrentLocation

	// Give the mastermind three "Increase Paranoia" cards (ID 4)
	card4Config, err := loader.Get[*v1.CardConfig](engine.gameConfig, 4)
	assert.NoError(t, err)
	mastermind.Hand = &v1.CardList{Cards: []*v1.Card{
		{Config: card4Config},
		{Config: card4Config},
		{Config: card4Config},
	}}

	// --- 执行：开始游戏并让其运行 ---
	engine.Start()
	defer engine.Stop()

	// 主谋打出所有三张牌
	for i := 0; i < 3; i++ {
		playCardAction := &v1.PlayerActionPayload{
			Payload: &v1.PlayerActionPayload_PlayCard{
				PlayCard: &v1.PlayCardPayload{
					CardId: 4,
					Target: &v1.PlayCardPayload_TargetCharacterId{TargetCharacterId: doctor.Config.Id},
				},
			},
		}
		engine.SubmitPlayerAction(mastermind.Id, playCardAction)
	}

	// --- 验证：等待一天结束，然后检查状态 ---
	waitForEvent(t, engine, v1.GameEventType_GAME_EVENT_TYPE_DAY_ADVANCED)
	engine.TriggerIncidents()                                                    // 为测试手动触发事件
	waitForEvent(t, engine, v1.GameEventType_GAME_EVENT_TYPE_INCIDENT_TRIGGERED) // 等待事件被处理

	doctorAfterIncident := engine.GetCharacterByID(2)
	assert.NotNil(t, doctorAfterIncident)
	assert.Equal(t, int32(3), doctorAfterIncident.Paranoia, "Paranoia should be 3")

	foundTrait := false
	for _, trait := range doctorAfterIncident.Traits {
		if trait == "Culprit" {
			foundTrait = true
			break
		}
	}
	assert.True(t, foundTrait, "The 'Culprit' trait should have been added to the Doctor")
}

func TestEngine_GetPlayerView(t *testing.T) {
	engine := helper_NewGameEngineForTest(t)
	engine.Start()
	defer engine.Stop()

	// --- 获取主谋和主角的视图 ---
	mastermindView := engine.GetPlayerView(1)  // 主谋 ID
	protagonistView := engine.GetPlayerView(2) // 主角 ID

	assert.NotNil(t, mastermindView)
	assert.NotNil(t, protagonistView)

	// --- 主谋验证 ---
	// 主谋应该能看到所有角色的隐藏角色。
	// 在“第一步”中，医生（ID 2）是“连环杀手”。
	doctorForMastermind := helper_GetCharacterFromView(t, mastermindView, 2)
	assert.NotNil(t, doctorForMastermind)
	assert.Equal(t, v1.RoleType_ROLE_TYPE_KILLER, doctorForMastermind.Role, "Mastermind should see the Doctor's hidden role")

	// --- 主角验证 ---
	// 主角不应该看到隐藏的角色。它应该是未知的。
	doctorForProtagonist := helper_GetCharacterFromView(t, protagonistView, 2)
	assert.NotNil(t, doctorForProtagonist)
	assert.Equal(t, v1.RoleType_ROLE_TYPE_ROLE_UNKNOWN, doctorForProtagonist.Role, "Protagonist should not see the Doctor's hidden role")

	// 两个玩家都应该能看到公共信息，比如角色的名字。
	assert.Equal(t, "Doctor", doctorForMastermind.Name)
	assert.Equal(t, "Doctor", doctorForProtagonist.Name)
}

func TestEngine_CharacterMovement(t *testing.T) {
	engine := helper_NewGameEngineForTest(t)
	char := engine.GetCharacterByID(1) // 高中女生，从学校开始
	assert.NotNil(t, char)
	assert.Equal(t, v1.LocationType_LOCATION_TYPE_SCHOOL, char.CurrentLocation)

	// --- 测试有效移动 ---
	// 从学校（1,0）向右移动到神社（0,0）- 这是水平移动，dx=-1，dy=0，但位置不是这样相邻的。
	// 让我们检查一下网格。学校是（1,0），神社是（0,0）。所以 dx 应该是 -1。
	// 让我们从学校（1,0）移动到城市（1,1），dy=1
	engine.MoveCharacter(char, 0, 1)
	assert.Equal(t, v1.LocationType_LOCATION_TYPE_CITY, char.CurrentLocation, "Should move from School to City")

	// 从城市（1,1）向左移动到医院（0,1），dx=-1
	engine.MoveCharacter(char, -1, 0)
	assert.Equal(t, v1.LocationType_LOCATION_TYPE_HOSPITAL, char.CurrentLocation, "Should move from City to Hospital")

	// --- 测试无效移动（越界）---
	// 尝试从医院（0,1）向左移动，越界
	engine.MoveCharacter(char, -1, 0)
	assert.Equal(t, v1.LocationType_LOCATION_TYPE_HOSPITAL, char.CurrentLocation, "Should not move out of bounds (left)")

	// 尝试从医院（0,1）向上移动到（0,0），即神社
	engine.MoveCharacter(char, 0, -1)
	assert.Equal(t, v1.LocationType_LOCATION_TYPE_SHRINE, char.CurrentLocation, "Should move from Hospital to Shrine")

	// 尝试从神社（0,0）向上移动，越界
	engine.MoveCharacter(char, 0, -1)
	assert.Equal(t, v1.LocationType_LOCATION_TYPE_SHRINE, char.CurrentLocation, "Should not move out of bounds (up)")
}

func TestEngine_GameOverOnMaxLoops(t *testing.T) {
	engine := helper_NewGameEngineForTest(t)

	// --- Setup: Deal cards to the mastermind ---
	mastermind := engine.GetMastermindPlayer()
	card4Config, err := loader.Get[*v1.CardConfig](engine.gameConfig, 4) // Increase Paranoia card
	assert.NoError(t, err)
	mastermind.Hand = &v1.CardList{Cards: []*v1.Card{
		{Config: card4Config},
		{Config: card4Config},
		{Config: card4Config},
	}}

	engine.Start()
	defer engine.Stop()

	// 手动将循环次数设置为最大值
	engine.GameState.CurrentLoop = engine.gameConfig.GetScript().LoopCount
	// 将日期设置为循环的最后一天
	engine.GameState.CurrentDay = engine.gameConfig.GetScript().DaysPerLoop

	// 等待游戏准备好玩家行动
	waitForPhase(t, engine, v1.GamePhase_GAME_PHASE_PROTAGONIST_CARD_PLAY)

	// --- 主谋出牌 ---
	playCardAction := &v1.PlayerActionPayload{
		Payload: &v1.PlayerActionPayload_PlayCard{
			PlayCard: &v1.PlayCardPayload{
				CardId: 4,
				Target: &v1.PlayCardPayload_TargetCharacterId{TargetCharacterId: 1},
			},
		},
	}
	for i := 0; i < 3; i++ {
		engine.SubmitPlayerAction(mastermind.Id, playCardAction)
	}

	// --- 主角跳过 ---
	passAction := &v1.PlayerActionPayload{
		Payload: &v1.PlayerActionPayload_PassTurn{
			PassTurn: &v1.PassTurnAction{},
		},
	}
	protagonists := engine.GetProtagonistPlayers()
	for _, player := range protagonists {
		engine.SubmitPlayerAction(player.Id, passAction)
	}

	// 我们期望一个 GAME_ENDED 事件，主角是赢家
	// 因为主谋未能在循环内实现其目标。
	waitForEvent(t, engine, v1.GameEventType_GAME_EVENT_TYPE_GAME_ENDED)

	// 如果需要，我们也可以检查最终阶段，但事件是更强的信号。
	assert.Equal(t, v1.GamePhase_GAME_PHASE_GAME_OVER, engine.GetCurrentPhase(), "Game should be in the GAME_OVER phasehandler")
}

// helper_GetCharacterFromView 是一个测试助手，用于通过 ID 在玩家视图中查找角色。
func helper_GetCharacterFromView(t *testing.T, view *v1.PlayerView, charID int32) *v1.PlayerViewCharacter {
	t.Helper()
	for _, char := range view.Characters {
		if char.Id == charID {
			return char
		}
	}
	t.Fatalf("Character with ID %d not found in player view", charID)
	return nil
}

// waitForEvent 是一个辅助函数，用于阻塞直到游戏引擎发出特定事件。
func waitForEvent(t *testing.T, engine *GameEngine, targetEvent v1.GameEventType) {
	t.Helper()
	timeout := time.After(2 * time.Second) // 2 秒超时

	// 我们需要从通道中消费事件以找到我们的目标事件。
	// 这应该在一个单独的 goroutine 中完成，以避免在事件永远不会到来时阻塞主测试线程。
	eventChan := engine.GetGameEvents()

	for {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for event %s", targetEvent)
		case event := <-eventChan:
			if event.Type == targetEvent {
				return
			}
		}
	}
}

// waitForPhase 是一个辅助函数，用于阻塞直到游戏引擎达到特定阶段。
func waitForPhase(t *testing.T, engine *GameEngine, targetPhase v1.GamePhase) {
	t.Helper()
	timeout := time.After(2 * time.Second)

	for {
		currentPhase := engine.GetCurrentPhase()
		if currentPhase == targetPhase {
			return
		}
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for phasehandler %s, current phasehandler is %s", targetPhase, currentPhase)
		case <-time.After(10 * time.Millisecond):
			// 继续轮询
		}
	}
}
