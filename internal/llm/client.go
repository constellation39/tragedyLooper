package llm

import (
	"encoding/json"
	"log"
	"time"
	model "tragedylooper/internal/game/proto/v1"

	"google.golang.org/protobuf/encoding/protojson"
)

// LLMClient 定义与 LLM 交互的接口。
type Client interface {
	GenerateResponse(prompt string, sessionID string) (string, error)
}

// MockLLMClient 是用于测试的模拟实现。
type MockLLMClient struct{}

func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{}
}

// GenerateResponse 模拟 LLM 响应。
func (m *MockLLMClient) GenerateResponse(prompt string, sessionID string) (string, error) {
	log.Printf("MockLLMClient: Received prompt for session %s:\n%s", sessionID, prompt)
	time.Sleep(500 * time.Millisecond) // 模拟 API 延迟

	// 简单的模拟逻辑：如果提示包含“play card”，则返回卡牌动作。
	// 实际中，LLM 将解析复杂的游戏状态并做出决策。
	if len(prompt) > 100 && prompt[len(prompt)-100:] == "Please provide your action in JSON format." {
		playCardPayload := &model.PlayCardPayload{
			CardId: 1, // 假设存在这张卡
			Target: &model.PlayCardPayload_TargetCharacterId{TargetCharacterId: 2},
		}
		// Marshal the payload to JSON
		jsonBytes, err := protojson.Marshal(playCardPayload)
		if err != nil {
			return "", err
		}

		// Unmarshal the JSON into a map
		var data map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &data); err != nil {
			return "", err
		}

		mockAction := model.PlayerActionPayload{
			Payload: &model.PlayerActionPayload_PlayCard{
				PlayCard: playCardPayload,
			},
		}
		actionBytes, _ := protojson.Marshal(&mockAction)
		return string(actionBytes), nil
	}

	return "Mock LLM response: I am thinking...", nil
}

/*
// OpenAIClient 是使用 OpenAI API 的示例实现。
type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
	}
}

func (o *OpenAIClient) GenerateResponse(prompt string, sessionID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo, // 或 openai.GPT4 等。
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an AI player in Tragedy Looper. Respond with a JSON object representing your action.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject, // 请求 JSON 输出
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("openai chat completion error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
*/
