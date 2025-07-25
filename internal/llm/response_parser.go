package llm

import (
	"encoding/json"
	"fmt"

	"tragedylooper/internal/game/model"
)

// ResponseParser 将 LLM 响应解析为结构化游戏动作。
type ResponseParser struct{}

func NewResponseParser() *ResponseParser {
	return &ResponseParser{}
}

// ParseLLMAction 将来自 LLM 的 JSON 字符串解析为 PlayerAction。
func (rp *ResponseParser) ParseLLMAction(llmResponse string) (model.PlayerAction, error) {
	var action model.PlayerAction
	err := json.Unmarshal([]byte(llmResponse), &action)
	if err != nil {
		return model.PlayerAction{}, fmt.Errorf("failed to unmarshal LLM response into PlayerAction: %w", err)
	}

	// 动作类型和负载结构的基本验证
	switch action.Type {
	case model.ActionPlayCard:
		var payload model.PlayCardPayload
		payloadBytes, err := json.Marshal(action.Payload)
		if err != nil {
			return model.PlayerAction{}, fmt.Errorf("failed to marshal PlayCard payload: %w", err)
		}
		err = json.Unmarshal(payloadBytes, &payload)
		if err != nil {
			return model.PlayerAction{}, fmt.Errorf("invalid PlayCard payload structure: %w", err)
		}
		action.Payload = payload // 替换为类型化负载

	case model.ActionUseAbility:
		var payload model.UseAbilityPayload
		payloadBytes, err := json.Marshal(action.Payload)
		if err != nil {
			return model.PlayerAction{}, fmt.Errorf("failed to marshal UseAbility payload: %w", err)
		}
		err = json.Unmarshal(payloadBytes, &payload)
		if err != nil {
			return model.PlayerAction{}, fmt.Errorf("invalid UseAbility payload structure: %w", err)
		}
		action.Payload = payload // 替换为类型化负载

	case model.ActionMakeGuess:
		var payload model.MakeGuessPayload
		payloadBytes, err := json.Marshal(action.Payload)
		if err != nil {
			return model.PlayerAction{}, fmt.Errorf("failed to marshal MakeGuess payload: %w", err)
		}
		err = json.Unmarshal(payloadBytes, &payload)
		if err != nil {
			return model.PlayerAction{}, fmt.Errorf("invalid MakeGuess payload structure: %w", err)
		}
		action.Payload = payload // 替换为类型化负载

	case model.ActionReadyForNextPhase:
		action.Payload = nil

	default:
		return model.PlayerAction{}, fmt.Errorf("unsupported action type from LLM: %s", action.Type)
	}

	return action, nil
}
