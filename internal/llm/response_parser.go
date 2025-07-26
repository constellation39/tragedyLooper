package llm

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
	"tragedylooper/internal/game/proto/model"
)

// ResponseParser 将 LLM 响应解析为结构化游戏动作。
type ResponseParser struct{}

func NewResponseParser() *ResponseParser {
	return &ResponseParser{}
}

// ParseLLMAction 将来自 LLM 的 JSON 字符串解析为 PlayerAction。
func (rp *ResponseParser) ParseLLMAction(llmResponse string) (*model.PlayerAction, error) {
	var action model.PlayerAction
	err := json.Unmarshal([]byte(llmResponse), &action)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal LLM response into PlayerAction: %w", err)
	}

	// 动作类型和负载结构的基本验证
	switch action.Type {
	case model.ActionType_ACTION_TYPE_PLAY_CARD:
		var payload model.PlayCardPayload
		payloadBytes, err := json.Marshal(action.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal PlayCard payload: %w", err)
		}
		err = json.Unmarshal(payloadBytes, &payload)
		if err != nil {
			return nil, fmt.Errorf("invalid PlayCard payload structure: %w", err)
		}
		action.Payload = &structpb.Struct{}
		if err := action.Payload.UnmarshalJSON(payloadBytes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload to struct: %w", err)
		}

	case model.ActionType_ACTION_TYPE_USE_ABILITY:
		var payload model.UseAbilityPayload
		payloadBytes, err := json.Marshal(action.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal UseAbility payload: %w", err)
		}
		err = json.Unmarshal(payloadBytes, &payload)
		if err != nil {
			return nil, fmt.Errorf("invalid UseAbility payload structure: %w", err)
		}
		action.Payload = &structpb.Struct{}
		if err := action.Payload.UnmarshalJSON(payloadBytes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload to struct: %w", err)
		}

	case model.ActionType_ACTION_TYPE_MAKE_GUESS:
		var payload model.MakeGuessPayload
		payloadBytes, err := json.Marshal(action.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal MakeGuess payload: %w", err)
		}
		err = json.Unmarshal(payloadBytes, &payload)
		if err != nil {
			return nil, fmt.Errorf("invalid MakeGuess payload structure: %w", err)
		}
		action.Payload = &structpb.Struct{}
		if err := action.Payload.UnmarshalJSON(payloadBytes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload to struct: %w", err)
		}

	case model.ActionType_ACTION_TYPE_READY_FOR_NEXT_PHASE:
		action.Payload = nil

	default:
		return nil, fmt.Errorf("unsupported action type from LLM: %s", action.Type)
	}

	return &action, nil
}
