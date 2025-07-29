package phase

import (
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// ProtagonistGuessPhase 主角猜测阶段
type ProtagonistGuessPhase struct{ basePhase }

// Type 返回阶段类型
func (p *ProtagonistGuessPhase) Type() model.GamePhase { return model.GamePhase_PROTAGONIST_GUESS }

// HandleAction 处理玩家操作
func (p *ProtagonistGuessPhase) HandleAction(ge GameEngine, playerID int32, action *model.PlayerActionPayload) Phase {
	state := ge.GetGameState()
	player, ok := state.Players[playerID]
	if !ok {
		ge.Logger().Warn("Action from unknown player", zap.Int32("playerID", playerID))
		return nil
	}

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_MakeGuess:
		return handleMakeGuessAction(ge, player, payload.MakeGuess)
	}
	return nil
}

func handleMakeGuessAction(ge GameEngine, player *model.Player, payload *model.MakeGuessPayload) Phase {
	// 目前，我们假设第一个猜测的主角结束游戏。
	if player.Role != model.PlayerRole_PROTAGONIST {
		ge.Logger().Warn("non-protagonist player tried to make a guess", zap.Int32("player_id", player.Id))
		return nil
	}

	script := ge.GetGameConfig().GetScript()
	if script == nil {
		ge.Logger().Error("failed to get script to verify guess")
		ge.ApplyAndPublishEvent(model.GameEventType_GAME_ENDED, &model.GameOverEvent{Winner: model.PlayerRole_MASTERMIND}) // 游戏结束，出现错误时主谋默认获胜
		return &GameOverPhase{}
	}

	correctGuesses := 0
	for _, roleInfo := range script.Characters {
		if guessedRole, ok := payload.GuessedRoles[roleInfo.CharacterId]; ok {
			if guessedRole == roleInfo.HiddenRole {
				correctGuesses++
			}
		}
	}

	if correctGuesses == len(script.Characters) {
		ge.ApplyAndPublishEvent(model.GameEventType_GAME_ENDED, &model.GameOverEvent{Winner: model.PlayerRole_PROTAGONIST})
	} else {
		ge.ApplyAndPublishEvent(model.GameEventType_GAME_ENDED, &model.GameOverEvent{Winner: model.PlayerRole_MASTERMIND})
	}
	return &GameOverPhase{}
}