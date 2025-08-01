package engine

import (
	"testing"
	"time"

	"github.com/constellation39/tragedyLooper/internal/game/loader"
	"github.com/constellation39/tragedyLooper/internal/logger"
	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"github.com/stretchr/testify/assert"
)

func helper_NewGameEngineForTest(t *testing.T) *GameEngine {
	t.Helper()

	log := logger.New()

	gameConfig, err := loader.LoadConfig("../../../data", "first_steps")
	if err != nil {
		t.Fatalf("failed to load game data: %v", err)
	}

	players := []*v1.Player{
		{Id: 1, Name: "Mastermind", Role: v1.PlayerRole_PLAYER_ROLE_MASTERMIND, IsLlm: false},
		{Id: 2, Name: "Protagonist 1", Role: v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, IsLlm: false},
		{Id: 3, Name: "Protagonist 2", Role: v1.PlayerRole_PLAYER_ROLE_PROTAGONIST, IsLlm: false},
	}

	engine, err := NewGameEngine(log, players, nil, gameConfig) // AI is not needed for these tests
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

	// --- Setup: Get characters and mastermind player ---
	mastermind := engine.GetMastermindPlayer()
	protagonist1 := engine.GetProtagonistPlayers()[0]
	protagonist2 := engine.GetProtagonistPlayers()[1]
	doctor := engine.GetCharacterByID(2)
	assert.NotNil(t, mastermind)
	assert.NotNil(t, doctor)

	// Manually give the mastermind three "Increase Paranoia" cards (ID 4)
	card4Config, err := loader.Get[*v1.CardConfig](engine.gameConfig, 4)
	assert.NoError(t, err)
	mastermind.Hand = &v1.CardList{Cards: []*v1.Card{
		{Config: card4Config},
		{Config: card4Config},
		{Config: card4Config},
	}}

	// --- Execution: Start the engine and play cards ---
	engine.Start()
	defer engine.Stop()

	passAction := &v1.PlayerActionPayload{
		Payload: &v1.PlayerActionPayload_PassTurn{
			PassTurn: &v1.PassTurnAction{},
		},
	}

	// The mastermind plays three "Increase Paranoia" cards on the Doctor over three days.
	for i := 0; i < 3; i++ {
		// Mastermind's turn
		engine.SubmitPlayerAction(mastermind.Id, &v1.PlayerActionPayload{
			Payload: &v1.PlayerActionPayload_PlayCard{
				PlayCard: &v1.PlayCardPayload{
					CardId: 4, // Increase Paranoia
					Target: &v1.PlayCardPayload_TargetCharacterId{TargetCharacterId: doctor.Config.Id},
				},
			},
		})

		// Protagonists' turn (they pass)
		engine.SubmitPlayerAction(protagonist1.Id, passAction)
		engine.SubmitPlayerAction(protagonist2.Id, passAction)

		// Let the engine process the turn. A small sleep is used here for simplicity,
		// but in a real-world scenario, we might wait for a specific phase transition.
		time.Sleep(50 * time.Millisecond)
	}

	// --- Verification: Wait for the incident to trigger and check the result ---
	// The "Culprit" incident should trigger when the Doctor's paranoia reaches 3.
	waitForEvent(t, engine, v1.GameEventType_GAME_EVENT_TYPE_INCIDENT_TRIGGERED)

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

	engine.Start()
	defer engine.Stop()

	// Manually set the loop to the max.
	// The game logic should detect this at the beginning of the loop
	// and transition directly to GAME_OVER.
	engine.GameState.CurrentLoop = engine.gameConfig.GetScript().GetLoopCount()

	// We expect a GAME_ENDED event because the loop count has been exceeded.
	waitForEvent(t, engine, v1.GameEventType_GAME_EVENT_TYPE_GAME_ENDED)

	// Optionally, we can also check the final phase, but the event is a stronger signal.
	assert.Equal(t, v1.GamePhase_GAME_PHASE_GAME_OVER, engine.GetCurrentPhase(), "Game should be in the GAME_OVER phase")
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
