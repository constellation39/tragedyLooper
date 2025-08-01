package phasehandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// ProtagonistGuessPhase is the phase where protagonists try to guess the hidden roles of other characters.
type ProtagonistGuessPhase struct {
	BasePhase
}

// Type 返回阶段类型，表示当前是主角猜测阶段。
func (p *ProtagonistGuessPhase) Type() model.GamePhase {
	return model.GamePhase_GAME_PHASE_PROTAGONIST_GUESS
}

// HandleAction 处理玩家在主角猜测阶段的操作。
func (p *ProtagonistGuessPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) bool {
	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_MakeGuess:
		// 目前，我们假设第一个猜测的主角结束游戏。
		if player.Role != model.PlayerRole_PLAYER_ROLE_PROTAGONIST {
			ge.Logger().Warn("non-protagonist player tried to make a guess", zap.Int32("player_id", player.Id))
			return false
		}

		script := ge.GetGameRepo().GetScript()
		if script == nil {
			ge.Logger().Error("failed to get script to verify guess")
			ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_GAME_ENDED, &model.EventPayload{
				Payload: &model.EventPayload_GameEnded{GameEnded: &model.GameEndedEvent{Winner: model.PlayerRole_PLAYER_ROLE_MASTERMIND, Reason: "Failed to guess all roles"}},
			}) // 游戏结束，出现错误时主谋默认获胜
			return true
		}

		correctGuesses := 0
		// 遍历剧本中的所有角色，检查猜测是否正确。
		for _, roleInfo := range script.Characters {
			if guessedRole, ok := payload.MakeGuess.GuessedRoles[roleInfo.CharacterId]; ok {
				if guessedRole == roleInfo.HiddenRole {
					correctGuesses++
				}
			}
		}

		// 如果所有猜测都正确，则主角获胜；否则主谋获胜。
		if correctGuesses == len(script.Characters) {
			ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_GAME_ENDED, &model.EventPayload{
				Payload: &model.EventPayload_GameEnded{GameEnded: &model.GameEndedEvent{Winner: model.PlayerRole_PLAYER_ROLE_PROTAGONIST, Reason: "Correctly guessed all roles"}},
			})
		} else {
			ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_GAME_ENDED, &model.EventPayload{
				Payload: &model.EventPayload_GameEnded{GameEnded: &model.GameEndedEvent{Winner: model.PlayerRole_PLAYER_ROLE_MASTERMIND, Reason: "Failed to guess all roles"}},
			})
		}
	case *model.PlayerActionPayload_PassTurn:
		ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_GAME_ENDED, &model.EventPayload{
			Payload: &model.EventPayload_GameEnded{GameEnded: &model.GameEndedEvent{Winner: model.PlayerRole_PLAYER_ROLE_MASTERMIND, Reason: "Failed to guess all roles"}},
		})
	}
	return true
}

func init() {
	RegisterPhase(&ProtagonistGuessPhase{})
}
