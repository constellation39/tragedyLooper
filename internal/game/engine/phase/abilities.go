package phase

import (
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// --- AbilitiesPhase ---
type AbilitiesPhase struct{ basePhase }

func (p *AbilitiesPhase) Type() model.GamePhase { return model.GamePhase_ABILITIES }
func (p *AbilitiesPhase) Enter(ge GameEngine) Phase {
	// Players can use abilities.
	// This phase might require player input and have a timeout.
	return &IncidentsPhase{}
}

func (p *AbilitiesPhase) HandleAction(ge GameEngine, playerID int32, action *model.PlayerActionPayload) Phase {
	state := ge.GetGameState()
	player, ok := state.Players[playerID]
	if !ok {
		ge.Logger().Warn("Action from unknown player", zap.Int32("playerID", playerID))
		return nil
	}

	switch payload := action.Payload.(type) {
	case *model.PlayerActionPayload_UseAbility:
		handleUseAbilityAction(ge, player, payload.UseAbility)
	}
	return nil
}

func handleUseAbilityAction(ge GameEngine, player *model.Player, payload *model.UseAbilityPayload) {
	var ability *model.Ability
	abilityFound := false
	char, ok := ge.GetGameState().Characters[payload.CharacterId]
	if !ok {
		ge.Logger().Warn("Character not found for ability use", zap.Int32("characterID", payload.CharacterId))
		return
	}

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

	// TODO: We need to re-implement applyEffect, as it was not part of the GameEngine interface.
	// if err := ge.applyEffect(ability.Config.Effect, ability, payload, nil); err != nil {
	// 	ge.Logger().Error("Failed to apply effect for ability", zap.String("abilityName", ability.Config.Name), zap.Error(err))
	// 	return
	// }

	if ability.Config.OncePerLoop {
		ability.UsedThisLoop = true
	}
	// Note: Using an ability does not automatically make a player "ready".
	// They must explicitly pass their turn with PassTurnAction.
}
