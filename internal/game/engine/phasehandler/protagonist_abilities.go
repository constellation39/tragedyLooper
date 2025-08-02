package phasehandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
	"go.uber.org/zap"
)

// ProtagonistAbilitiesPhase is the phase where protagonists can use character abilities.
type ProtagonistAbilitiesPhase struct {
	BasePhase
	protagonistTurnIndex int
}

// Type returns the phase type.
func (p *ProtagonistAbilitiesPhase) Type() model.GamePhase {
	return model.GamePhase_GAME_PHASE_PROTAGONIST_ABILITIES
}

// Enter is called at the beginning of the phase.
func (p *ProtagonistAbilitiesPhase) Enter(ge GameEngine) PhaseState {
	p.protagonistTurnIndex = 0

	// If no protagonists need to act, move to the next phase.
	if len(ge.GetProtagonistPlayers()) == 0 {
		return PhaseComplete
	}

	// Trigger AI for the first protagonist if applicable.
	// ge.RequestAIAction(ge.GetProtagonistPlayers()[0].Id)
	return PhaseInProgress
}

// HandleAction handles actions from the player.
func (p *ProtagonistAbilitiesPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) PhaseState {
	if !p.isActionInTurn(ge, player) {
		ge.Logger().Warn("Received action from player out of turn", zap.String("player", player.Name))
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
func (p *ProtagonistAbilitiesPhase) HandleTimeout(ge GameEngine) {
	ge.Logger().Info("Protagonist abilities phase timed out, passing turn.")
	p.handlePassTurn(ge)
}

func (p *ProtagonistAbilitiesPhase) isActionInTurn(ge GameEngine, player *model.Player) bool {
	protagonists := ge.GetProtagonistPlayers()
	if p.protagonistTurnIndex >= len(protagonists) {
		return false // Should not happen
	}
	return player.Id == protagonists[p.protagonistTurnIndex].Id
}

func (p *ProtagonistAbilitiesPhase) handlePassTurn(ge GameEngine) PhaseState {
	p.protagonistTurnIndex++
	protagonists := ge.GetProtagonistPlayers()
	if p.protagonistTurnIndex >= len(protagonists) {
		ge.Logger().Info("All protagonists have acted, moving to Incidents Phase")
		return PhaseComplete
	}

	// Trigger AI for the next protagonist if applicable.
	// ge.RequestAIAction(protagonists[p.protagonistTurnIndex].Id)
	return PhaseInProgress
}

func (p *ProtagonistAbilitiesPhase) handleUseAbilityAction(ge GameEngine, player *model.Player, payload *model.UseAbilityPayload) {
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

func init() {
	RegisterPhase(&ProtagonistAbilitiesPhase{})
}
