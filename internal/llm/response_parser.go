package llm

import (
	"encoding/json"
	"fmt"

	model "tragedylooper/internal/game/proto/v1"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
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
		if err := unmarshalPayload(action.Payload, &payload); err != nil {
			return nil, err
		}
		anyPayload, err := anypb.New(&payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal PlayCard payload to any: %w", err)
		}
		action.Payload = anyPayload

	case model.ActionType_ACTION_TYPE_USE_ABILITY:
		var payload model.UseAbilityPayload
		if err := unmarshalPayload(action.Payload, &payload); err != nil {
			return nil, err
		}
		anyPayload, err := anypb.New(&payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal UseAbility payload to any: %w", err)
		}
		action.Payload = anyPayload

	case model.ActionType_ACTION_TYPE_MAKE_GUESS:
		var payload model.MakeGuessPayload
		if err := unmarshalPayload(action.Payload, &payload); err != nil {
			return nil, err
		}
		anyPayload, err := anypb.New(&payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal MakeGuess payload to any: %w", err)
		}
		action.Payload = anyPayload

	case model.ActionType_ACTION_TYPE_READY_FOR_NEXT_PHASE:
		action.Payload = nil

	default:
		return nil, fmt.Errorf("unsupported action type from LLM: %s", action.Type)
	}

	return &action, nil
}

func unmarshalPayload(any *anypb.Any, msg proto.Message) error {
	if any == nil {
		return fmt.Errorf("payload is nil")
	}
	// First, unmarshal the Any into a structpb.Struct
	structPayload := &structpb.Struct{}
	if err := any.UnmarshalTo(structPayload); err != nil {
		// Fallback for older format
		return any.UnmarshalTo(msg)
	}

	// Then, marshal the struct to JSON
	jsonBytes, err := protojson.Marshal(structPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal struct to json: %w", err)
	}

	// Finally, unmarshal the JSON into the target message
	return protojson.Unmarshal(jsonBytes, msg)
}
