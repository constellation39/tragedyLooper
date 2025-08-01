package llm

import (
	"context"
	"fmt"

	"github.com/constellation39/tragedyLooper/internal/game/engine/ai"

	"go.uber.org/zap"

	promptbuilder "github.com/constellation39/tragedyLooper/internal/llm/prompt"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// LLMActionGenerator is an implementation of engine.ActionGenerator that uses an LLM.
type LLMActionGenerator struct {
	Client Client
	Logger *zap.Logger
}

// NewLLMActionGenerator creates a new LLMActionGenerator.
func NewLLMActionGenerator(client Client, logger *zap.Logger) *LLMActionGenerator {
	return &LLMActionGenerator{Client: client, Logger: logger}
}

// GenerateAction prompts the LLM to generate a player action based on the provided context.
func (g *LLMActionGenerator) GenerateAction(ctx context.Context, data *ai.ActionGeneratorContext) (*model.PlayerActionPayload, error) {
	g.Logger.Info("Generating action for LLM player", zap.String("player", data.Player.Name))

	pBuilder := promptbuilder.NewPromptBuilder()
	var prompt string

	if data.Player.Role == model.PlayerRole_PLAYER_ROLE_MASTERMIND {
		charactersWithStringKeys := make(map[string]*model.Character)
		for id, char := range data.AllCharacters {
			charactersWithStringKeys[fmt.Sprint(id)] = char
		}
		prompt = pBuilder.BuildMastermindPrompt(data.PlayerView, data.Script, charactersWithStringKeys)
	} else {
		deductionKnowledgeWithStringKeys := make(map[string]string)
		for id, value := range data.Player.DeductionKnowledge.GuessedRoles {
			deductionKnowledgeWithStringKeys[fmt.Sprint(id)] = value.String()
		}
		prompt = pBuilder.BuildProtagonistPrompt(data.PlayerView, deductionKnowledgeWithStringKeys)
	}

	llmResponse, err := g.Client.GenerateResponse(prompt, data.Player.LlmSessionId)
	if err != nil {
		g.Logger.Error("LLM call failed", zap.String("player", data.Player.Name), zap.Error(err))
		return nil, fmt.Errorf("llm call failed for player %s: %w", data.Player.Name, err)
	}

	responseParser := NewResponseParser()
	llmAction, err := responseParser.ParseLLMAction(llmResponse)
	if err != nil {
		g.Logger.Error("Failed to parse LLM response", zap.String("player", data.Player.Name), zap.Error(err))
		return nil, fmt.Errorf("failed to parse llm response for player %s: %w", data.Player.Name, err)
	}

	return llmAction, nil
}
