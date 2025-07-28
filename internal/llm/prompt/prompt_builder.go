package promptbuilder

import (
	"encoding/json"
	"fmt"
	"strings"

	model "tragedylooper/internal/game/proto/v1"
)

// PromptBuilder 帮助构建 LLM 玩家的提示词。
type PromptBuilder struct{}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// BuildMastermindPrompt 为主谋 LLM 构建提示词。
func (pb *PromptBuilder) BuildMastermindPrompt(
	fullGameState *model.PlayerView, // 主谋获得完整视图
	script *model.ScriptConfig,
	characters map[string]*model.Character, // 主谋看到隐藏身份
) string {
	var sb strings.Builder
	sb.WriteString("You are the Mastermind in the model Tragedy Looper.\n")
	sb.WriteString("Your goal is to trigger one of the tragedies defined in the script before the loops run out, or before the Protagonists make a correct final guess.\n")
	sb.WriteString("You know all hidden roles and plots. You can bluff and mislead the Protagonists.\n\n")

	sb.WriteString("--- Game State ---\n")
	sb.WriteString(fmt.Sprintf("Current Loop: %d/%d, Current Day: %d/%d\n", fullGameState.CurrentLoop, script.LoopCount, fullGameState.CurrentDay, script.DaysPerLoop))
	sb.WriteString(fmt.Sprintf("Current Phase: %s\n", fullGameState.CurrentPhase))

	sb.WriteString("\n--- Script Details ---\n")
	sb.WriteString(fmt.Sprintf("Script Name: %s\n", script.Name))

	// TODO: add subplots to prompt
	// if len(script.SubPlots) > 0 {
	//	sb.WriteString(fmt.Sprintf("Sub Plots: %s\n", strings.Join(script.SubPlots, ", ")))
	//}
	sb.WriteString("Tragedies to trigger:\n")
	for _, t := range script.Incidents {
		sb.WriteString(fmt.Sprintf("- %s (Day %d, Culprit: %d, Conditions: %+v)\n", t.IncidentType, t.Day, t.CulpritCharacterId, t.TriggerConditions))
	}

	sb.WriteString("\n--- Characters (including hidden roles) ---\n")
	for _, char := range characters { // 使用主谋的完整角色映射
		sb.WriteString(fmt.Sprintf("- %s (Role: %s, Location: %s, Paranoia: %d, Goodwill: %d, Alive: %t)\n",
			char.Config.Name, char.HiddenRole, char.CurrentLocation, char.Paranoia, char.Goodwill, char.IsAlive))
		if len(char.Config.Traits) > 0 {
			sb.WriteString(fmt.Sprintf("  Traits: %s\n", strings.Join(char.Config.Traits, ", ")))
		}
	}

	sb.WriteString("\n--- Your Hand ---\n")
	for _, card := range fullGameState.YourHand {
		sb.WriteString(fmt.Sprintf("- Card: %s (Type: %s, Effect: %+v)\n", card.Config.Name, card.Config.CardType, card.Config.Effect))
	}

	sb.WriteString("\n--- Public Events (This Day) ---\n")
	for _, event := range fullGameState.PublicEvents {
		eventBytes, _ := json.Marshal(event.Payload)
		sb.WriteString(fmt.Sprintf("- %s: %s\n", event.Type, string(eventBytes)))
	}

	sb.WriteString("\n--- Instructions ---\n")
	sb.WriteString(fmt.Sprintf("It is currently the %s phase.\n", fullGameState.CurrentPhase))
	sb.WriteString("Based on the current model state and your objective, decide your action.\n")
	sb.WriteString("You can play a card (ActionPlayCard) or use an ability (ActionUseAbility) or signal readiness (ActionReadyForNextPhase).\n")
	sb.WriteString("If playing a card, specify 'card_id', 'target_character_id' (if applicable), 'target_location' (if applicable).\n")
	sb.WriteString("If using an ability, specify 'ability_name' and 'target_character_id' (if applicable).\n")
	sb.WriteString("If you have no action or are done, use ActionReadyForNextPhase.\n")
	sb.WriteString("Please provide your action in JSON format, matching the model.PlayerAction structure.\n")
	sb.WriteString("Example: {\"type\": \"PlayCard\", \"payload\": {\"card_id\": \"some_card_id\", \"target_character_id\": \"some_char_id\"}}\n")
	sb.WriteString("Example: {\"type\": \"ReadyForNextPhase\"}\n")

	return sb.String()
}

// BuildProtagonistPrompt 为主角 LLM 构建提示词。
func (pb *PromptBuilder) BuildProtagonistPrompt(
	playerView *model.PlayerView,
	deductionKnowledge map[string]string,
) string {
	var sb strings.Builder
	sb.WriteString("You are a Protagonist in the model Tragedy Looper.\n")
	sb.WriteString("Your goal is to deduce the hidden roles and plots, and prevent tragedies from occurring.\n")
	sb.WriteString("You do not know the hidden roles or the full script initially. You must deduce them from observations.\n\n")

	sb.WriteString("--- Game State ---\n")
	sb.WriteString(fmt.Sprintf("Current Loop: %d, Current Day: %d\n", playerView.CurrentLoop, playerView.CurrentDay))
	sb.WriteString(fmt.Sprintf("Current Phase: %s\n", playerView.CurrentPhase))

	sb.WriteString("\n--- Characters (visible information) ---\n")
	for _, char := range playerView.Characters { // 主角视图中隐藏了隐藏身份
		sb.WriteString(fmt.Sprintf("- %s (Location: %s, Paranoia: %d, Goodwill: %d, Alive: %t)\n",
			char.Name, char.CurrentLocation, char.Paranoia, char.Goodwill, char.IsAlive))
		if len(char.Traits) > 0 {
			sb.WriteString(fmt.Sprintf("  Traits: %s\n", strings.Join(char.Traits, ", ")))
		}
	}

	sb.WriteString("\n--- Your Hand ---\n")
	for _, card := range playerView.YourHand {
		sb.WriteString(fmt.Sprintf("- Card: %s (Type: %s, Effect: %+v)\n", card.Config.Name, card.Config.CardType, card.Config.Effect))
	}

	sb.WriteString("\n--- Your Deductions (from previous loops) ---\n")
	if len(deductionKnowledge) > 0 {
		deductionBytes, _ := json.MarshalIndent(deductionKnowledge, "", "  ")
		sb.WriteString(string(deductionBytes))
	} else {
		sb.WriteString("No deductions yet.\n")
	}

	sb.WriteString("\n--- Public Events (This Day) ---\n")
	for _, event := range playerView.PublicEvents {
		eventBytes, _ := json.Marshal(event.Payload)
		sb.WriteString(fmt.Sprintf("- %s: %s\n", event.Type, string(eventBytes)))
	}

	sb.WriteString("\n--- Instructions ---\n")
	sb.WriteString(fmt.Sprintf("It is currently the %s phase.\n", playerView.CurrentPhase))
	sb.WriteString("Based on the current model state and your deductions, decide your action.\n")
	sb.WriteString("You can play a card (ActionPlayCard) or use an ability (ActionUseAbility) or signal readiness (ActionReadyForNextPhase).\n")
	sb.WriteString("If playing a card, specify 'card_id', 'target_character_id' (if applicable), 'target_location' (if applicable).\n")
	sb.WriteString("If using an ability, specify 'ability_name' and 'target_character_id' (if applicable).\n")
	sb.WriteString("If you have no action or are done, use ActionReadyForNextPhase.\n")
	sb.WriteString("You can also make a final guess (ActionMakeGuess) if you believe you have deduced all roles correctly. This ends the model.\n")
	sb.WriteString("Please provide your action in JSON format, matching the model.PlayerAction structure.\n")
	sb.WriteString("Example: {\"type\": \"PlayCard\", \"payload\": {\"card_id\": \"some_card_id\", \"target_character_id\": \"some_char_id\"}}\n")
	sb.WriteString("Example: {\"type\": \"ReadyForNextPhase\"}\n")
	sb.WriteString("Example: {\"type\": \"MakeGuess\", \"payload\": {\"guessed_roles\": {\"char_id_1\": \"Killer\", \"char_id_2\": \"Brain\"}}}\n")

	return sb.String()
}