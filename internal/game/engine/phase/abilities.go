package phase

import (
	"time"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

type AbilityTurn int

const (
	MastermindAbilityTurn AbilityTurn = iota
	ProtagonistAbilityTurn
)

// AbilitiesPhase is the phase where players can use character abilities.
type AbilitiesPhase struct {
	basePhase
	turn                 AbilityTurn
	protagonistTurnIndex int
}

// Type returns the phase type.
func (p *AbilitiesPhase) Type() model.GamePhase { return model.GamePhase_ABILITIES }

// Enter is called when the phase begins.
func (p *AbilitiesPhase) Enter(ge GameEngine) Phase {
	p.turn = MastermindAbilityTurn
	p.protagonistTurnIndex = 0
	ge.ResetPlayerReadiness()

	// Optional: Trigger AI for mastermind
	// ge.TriggerAIPlayerAction(ge.GetMastermindPlayer().Id)

	return nil
}

// HandleAction handles an action from a player.
func (p *AbilitiesPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if !p.isActionInTurn(ge, player) {
		ge.Logger().Warn("Received action from player out of turn", zap.String("player", player.Name))
		return nil
	}

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_UseAbility:
		p.handleUseAbilityAction(ge, player, payload.UseAbility)
		// Note: Using an ability does not automatically end the turn.
		// The player must explicitly pass.
	case *model.PlayerActionPayload_PassTurn:
		return p.handlePassTurn(ge, player)
	}

	return nil
}

// HandleTimeout handles a timeout.
func (p *AbilitiesPhase) HandleTimeout(ge GameEngine) Phase {
	ge.Logger().Info("Abilities phase timed out, passing turn.")
	var player *model.Player
	if p.turn == MastermindAbilityTurn {
		player = ge.GetMastermindPlayer()
	} else {
		protagonists := ge.GetProtagonistPlayers()
		if p.protagonistTurnIndex < len(protagonists) {
			player = protagonists[p.protagonistTurnIndex]
		}
	}
	if player != nil {
		return p.handlePassTurn(ge, player)
	}
	return &IncidentsPhase{}
}

// TimeoutDuration returns the timeout duration for this phase.
func (p *AbilitiesPhase) TimeoutDuration() time.Duration { return 60 * time.Second }

func (p *AbilitiesPhase) isActionInTurn(ge GameEngine, player *model.Player) bool {
	if p.turn == MastermindAbilityTurn {
		return player.Role == model.PlayerRole_MASTERMIND
	}

	protagonists := ge.GetProtagonistPlayers()
	if p.protagonistTurnIndex >= len(protagonists) {
		return false // Should not happen
	}
	return player.Id == protagonists[p.protagonistTurnIndex].Id
}

func (p *AbilitiesPhase) handlePassTurn(ge GameEngine, player *model.Player) Phase {
	ge.Logger().Info("Player passed ability turn", zap.String("player", player.Name))

	if p.turn == MastermindAbilityTurn {
		p.turn = ProtagonistAbilityTurn
		ge.Logger().Info("Transitioning to Protagonist ability turn")
		// Optional: Trigger AI for the first protagonist
		// protagonists := ge.GetProtagonistPlayers()
		// if len(protagonists) > 0 {
		// 	ge.TriggerAIPlayerAction(protagonists[0].Id)
		// }
		return nil
	}

	p.protagonistTurnIndex++
	protagonists := ge.GetProtagonistPlayers()
	if p.protagonistTurnIndex >= len(protagonists) {
		ge.Logger().Info("All protagonists have acted, moving to Incidents Phase")
		return &IncidentsPhase{}
	}

	// Optional: Trigger AI for the next protagonist
	// ge.TriggerAIPlayerAction(protagonists[p.protagonistTurnIndex].Id)
	return nil
}

func (p *AbilitiesPhase) handleUseAbilityAction(ge GameEngine, player *model.Player, payload *model.UseAbilityPayload) {
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

	// Here we need to check if the player has the right to use this ability.
	// For now, we assume if they are in turn, they can.
	// A more complex check for Goodwill abilities might be needed here.

	if err := ge.ApplyEffect(ability.Config.Effect, ability, payload, nil); err != nil {
		ge.Logger().Error("Failed to apply effect for ability", zap.String("abilityName", ability.Config.Name), zap.Error(err))
		return
	}

	if ability.Config.OncePerLoop {
		ability.UsedThisLoop = true
	}

	ge.Logger().Info("Player used ability", zap.String("player", player.Name), zap.String("ability", ability.Config.Name))
}
