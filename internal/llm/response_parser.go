package llm

import (
	"encoding/json"
	"fmt"

	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// ResponseParser 将 LLM 响应解析为结构化游戏动作。
type ResponseParser struct{}

func NewResponseParser() *ResponseParser {
	return &ResponseParser{}
}

// ParseLLMAction 将来自 LLM 的 JSON 字符串解析为 PlayerAction。
func (rp *ResponseParser) ParseLLMAction(llmResponse string) (*model.PlayerActionPayload, error) {
	var action model.PlayerActionPayload
	err := json.Unmarshal([]byte(llmResponse), &action)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal LLM response into PlayerAction: %w", err)
	}

	// 动作类型和负载结构的基本验证
	switch action.Payload.(type) {
	case *model.PlayerActionPayload_PlayCard:
		// No further processing needed for now

	case *model.PlayerActionPayload_UseAbility:
		// No further processing needed for now

	case *model.PlayerActionPayload_MakeGuess:
		// No further processing needed for now

	case *model.PlayerActionPayload_ChooseOption:
		// No further processing needed for now

	default:
		return nil, fmt.Errorf("unsupported action type from LLM: %T", action.Payload)
	}

	return &action, nil
}
