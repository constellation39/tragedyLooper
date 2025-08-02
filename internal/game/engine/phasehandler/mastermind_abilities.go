package phasehandler

import (
	"github.com/constellation39/tragedyLooper/internal/game/engine/condition"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
	"go.uber.org/zap"
)

// MastermindAbilitiesPhase is the phase where the mastermind can use character abilities.
type MastermindAbilitiesPhase struct {
	BasePhase
}

// Type returns the phase type.
func (p *MastermindAbilitiesPhase) Type() model.GamePhase {
	return model.GamePhase_GAME_PHASE_MASTERMIND_ABILITIES
}

// Enter is called at the beginning of the phase.
func (p *MastermindAbilitiesPhase) Enter(ge GameEngine) PhaseState {
	// Check for any events triggered by actions from the previous phase.
	checkTriggers(ge)

	// If the mastermind doesn't need to act, move to the next phase.
	if ge.GetMastermindPlayer() == nil {
		return PhaseComplete
	}

	// Trigger AI for the mastermind if applicable.
	ge.RequestAIAction(ge.GetMastermindPlayer().Id)
	return PhaseInProgress
}

// HandleAction handles actions from the player.
func (p *MastermindAbilitiesPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) PhaseState {
	if player.Role != model.PlayerRole_PLAYER_ROLE_MASTERMIND {
		ge.Logger().Warn("Received action from non-mastermind player during mastermind abilities phase", zap.String("player", player.Name))
		return PhaseInProgress
	}

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_UseAbility:
		p.handleUseAbilityAction(ge, player, payload.UseAbility)
	case *model.PlayerActionPayload_PassTurn:
		return p.handlePassTurn(ge)
	}
	return PhaseInProgress
}

// HandleTimeout handles a timeout.
func (p *MastermindAbilitiesPhase) HandleTimeout(ge GameEngine) {
	ge.Logger().Info("Mastermind abilities phase timed out, passing turn.")
	p.handlePassTurn(ge)
}

func (p *MastermindAbilitiesPhase) handlePassTurn(ge GameEngine) PhaseState {
	ge.Logger().Info("Mastermind passed ability turn, moving to Protagonist Abilities Phase")
	return PhaseComplete
}

func (p *MastermindAbilitiesPhase) handleUseAbilityAction(ge GameEngine, player *model.Player, payload *model.UseAbilityPayload) {
	char, ok := ge.GetGameState().Characters[payload.CharacterId]
	if !ok {
		ge.Logger().Warn("Character not found for ability use", zap.Int32("characterID", payload.CharacterId))
		return
	}

	var ability *model.Ability
	abilityFound := false
	for i := range char.Abilities {
		if char.Abilities[i].Config.Id == payload.AbilityId {
			ability = char.Abilities[i]
			abilityFound = true
			break
		}
	}

	if !abilityFound {
		ge.Logger().Warn("Ability not found on character", zap.Int32("abilityID", payload.AbilityId), zap.Int32("characterID", payload.CharacterId))
		return
	}

	if ability.UsedThisLoop {
		ge.Logger().Warn("Ability has already been used this loop", zap.String("abilityName", ability.Config.Name))
		return
	}

	if err := ge.ApplyEffect(ability.Config.Effect, ability, payload, nil); err != nil {
		ge.Logger().Error("Failed to apply effect for ability", zap.String("abilityName", ability.Config.Name), zap.Error(err))
		return
	}

	if ability.Config.OncePerLoop {
		ability.UsedThisLoop = true
	}

	ge.Logger().Info("Player used ability", zap.String("player", player.Name), zap.String("ability", ability.Config.Name))
}

func checkTriggers(ge GameEngine) {
	gs := ge.GetGameState()

	// 1. Check for incident triggers
	for _, incident := range ge.GetGameRepo().GetIncidents() {
		// Skip already triggered incidents
		if gs.TriggeredIncidents[incident.GetName()] {
			continue
		}

		triggerConditions := incident.GetTriggerConditions()
		if len(triggerConditions) == 0 {
			continue
		}

		compoundCondition := &model.Condition{
			ConditionType: &model.Condition_CompoundCondition{
				CompoundCondition: &model.CompoundCondition{
					Operator:      model.CompoundCondition_OPERATOR_AND,
					SubConditions: triggerConditions,
				},
			},
		}
		triggered, err := condition.Check(ge.GetGameState(), compoundCondition)
		if err != nil {
			ge.Logger().Error("Error checking incident trigger", zap.String("incident", incident.GetName()), zap.Error(err))
			continue
		}

		if triggered {
			ge.Logger().Info("Incident triggered", zap.String("incident", incident.GetName()))
			gs.TriggeredIncidents[incident.GetName()] = true
			ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_INCIDENT_TRIGGERED, &model.EventPayload{
				Payload: &model.EventPayload_IncidentTriggered{IncidentTriggered: &model.IncidentTriggeredEvent{Incident: &model.Incident{
					Config:               incident,
					Name:                 "",
					Day:                  0,
					Culprit:              "",
					Victim:               "",
					Description:          "",
					HasTriggeredThisLoop: false,
				}}},
			})
		}
	}

	// 2. Check for tragedy triggers (if applicable)
	// ... tragedy check logic should be added here ...

	// 3. Check for game over conditions
	checkEndConditions(ge)
}

// checkEndConditions checks if any game-ending conditions are met.
func checkEndConditions(ge GameEngine) {
	gs := ge.GetGameState()
	script := ge.GetGameRepo().GetScript()

	// Check if the maximum number of loops has been reached
	if gs.CurrentLoop >= script.GetLoopCount() {
		ge.Logger().Info("Max loops reached. Protagonists win.")
		ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_GAME_ENDED, &model.EventPayload{
			Payload: &model.EventPayload_GameEnded{GameEnded: &model.GameEndedEvent{Winner: model.PlayerRole_PLAYER_ROLE_PROTAGONIST}},
		})
	}

	// ... other game-ending conditions can be added here (e.g., all protagonists are dead) ...
}

func init() {
	RegisterPhase(&MastermindAbilitiesPhase{})
}
