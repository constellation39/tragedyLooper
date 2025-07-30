package phase

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// ProtagonistGuessPhase 主角猜测阶段，主角在此阶段尝试猜测其他角色的隐藏身份。
type ProtagonistGuessPhase struct{ basePhase }

// Type 返回阶段类型，表示当前是主角猜测阶段。
func (p *ProtagonistGuessPhase) Type() model.GamePhase { return model.GamePhase_PROTAGONIST_GUESS }

// HandleAction 处理玩家在主角猜测阶段的操作。
func (p *ProtagonistGuessPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if payload, ok := action.Payload.(*model.PlayerActionPayload_MakeGuess); ok {
		return handleMakeGuessAction(ge, player, payload.MakeGuess)
	}
	return nil
}

// handleMakeGuessAction 处理玩家进行猜测的操作。
func handleMakeGuessAction(ge GameEngine, player *model.Player, payload *model.MakeGuessPayload) Phase {
	// 目前，我们假设第一个猜测的主角结束游戏。
	if player.Role != model.PlayerRole_PROTAGONIST {
		ge.Logger().Warn("non-protagonist player tried to make a guess", zap.Int32("player_id", player.Id))
		return nil
	}

	script := ge.GetGameRepo().GetScript()
	if script == nil {
		ge.Logger().Error("failed to get script to verify guess")
		ge.ApplyAndPublishEvent(model.GameEventType_GAME_ENDED, &model.EventPayload{
			Payload: &model.EventPayload_GameOver{GameOver: &model.GameOverEvent{Winner: model.PlayerRole_MASTERMIND}},
		}) // 游戏结束，出现错误时主谋默认获胜
		return &GameOverPhase{}
	}

	correctGuesses := 0
	// 遍历剧本中的所有角色，检查猜测是否正确。
	for _, roleInfo := range script.Characters {
		if guessedRole, ok := payload.GuessedRoles[roleInfo.CharacterId]; ok {
			if guessedRole == roleInfo.HiddenRole {
				correctGuesses++
			}
		}
	}

	// 如果所有猜测都正确，则主角获胜；否则主谋获胜。
	if correctGuesses == len(script.Characters) {
		ge.ApplyAndPublishEvent(model.GameEventType_GAME_ENDED, &model.EventPayload{
			Payload: &model.EventPayload_GameOver{GameOver: &model.GameOverEvent{Winner: model.PlayerRole_PROTAGONIST}},
		})
	} else {
		ge.ApplyAndPublishEvent(model.GameEventType_GAME_ENDED, &model.EventPayload{
			Payload: &model.EventPayload_GameOver{GameOver: &model.GameOverEvent{Winner: model.PlayerRole_MASTERMIND}},
		})
	}
	return &GameOverPhase{}
}
