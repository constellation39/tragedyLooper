package engine

import (
	"fmt"

	"go.uber.org/zap"

	model "tragedylooper/internal/game/proto/v1"
	"tragedylooper/internal/llm"
	promptbuilder "tragedylooper/internal/llm/prompt"
)

// --- LLM Integration ---

// triggerLLMPlayerAction prompts an LLM player to make a decision.
func (ge *GameEngine) triggerLLMPlayerAction(playerID int32) {
	player := ge.GameState.Players[playerID]
	if player == nil || !player.IsLlm {
		return
	}

	ge.logger.Info("Triggering LLM for player", zap.String("player", player.Name), zap.String("role", string(player.Role)))
	playerView := ge.GetPlayerView(playerID) // Get a safe, filtered view of the game state
	pBuilder := promptbuilder.NewPromptBuilder()
	var prompt string
	if player.Role == model.PlayerRole_MASTERMIND {
		charactersWithStringKeys := make(map[string]*model.Character)
		for id, char := range ge.GameState.Characters {
			charactersWithStringKeys[fmt.Sprint(id)] = char
		}
		prompt = pBuilder.BuildMastermindPrompt(playerView, ge.GameState.Script, charactersWithStringKeys)
	} else {
		deductionKnowledgeWithStringKeys := make(map[string]string)
		for id, value := range player.DeductionKnowledge.GuessedRoles {
			deductionKnowledgeWithStringKeys[fmt.Sprint(id)] = value.String()
		}
		prompt = pBuilder.BuildProtagonistPrompt(playerView, deductionKnowledgeWithStringKeys)
	}

	go func() {
		llmResponse, err := ge.llmClient.GenerateResponse(prompt, player.LlmSessionId)
		if err != nil {
			ge.logger.Error("LLM call failed", zap.String("player", player.Name), zap.Error(err))
			// Submit a default action to unblock the game
			ge.requestChan <- &llmActionCompleteRequest{
				playerID: playerID,
				action:   &model.PlayerActionPayload{},
			}
			return
		}

		responseParser := llm.NewResponseParser()
		llmAction, err := responseParser.ParseLLMAction(llmResponse)
		if err != nil {
			ge.logger.Error("Failed to parse LLM response", zap.String("player", player.Name), zap.Error(err))
			// Submit a default action to unblock the game
			ge.requestChan <- &llmActionCompleteRequest{
				playerID: playerID,
				action:   &model.PlayerActionPayload{},
			}
			return
		}

		// Here, a symbolic AI component could validate or refine the LLM's suggestion.
		// This is the core of the "Hybrid AI" approach. For now, we trust the LLM's action.

		// Send the validated action back to the main loop for processing.
		ge.requestChan <- &llmActionCompleteRequest{
			playerID: playerID,
			action:   llmAction,
		}
	}()
}
