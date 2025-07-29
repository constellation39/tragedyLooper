package engine

import (
	"time"
	"tragedylooper/internal/game/loader"
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

func (ge *GameEngine) getProtagonistPlayers() []*model.Player {
	var protagonists []*model.Player
	for _, p := range ge.GameState.Players {
		if p.Role == model.PlayerRole_PROTAGONIST {
			protagonists = append(protagonists, p)
		}
	}
	return protagonists
}

func (ge *GameEngine) getMastermindPlayer() *model.Player {
	for _, p := range ge.GameState.Players {
		if p.Role == model.PlayerRole_MASTERMIND {
			return p
		}
	}
	return nil
}

func (ge *GameEngine) isMastermindReady() bool {
	mm := ge.getMastermindPlayer()
	if mm == nil {
		return true // No mastermind, so they are vacuously ready
	}
	return ge.playerReady[mm.Id]
}

func (ge *GameEngine) areAllPlayersReady() bool {
	for _, p := range ge.GameState.Players {
		if !ge.playerReady[p.Id] {
			return false
		}
	}
	return true
}

func (ge *GameEngine) checkConditions(conditions []*model.Condition, player *model.Player, payload *model.UseAbilityPayload, ability *model.Ability) bool {
	return true
}

func (ge *GameEngine) checkGameEndConditions() (bool, model.PlayerRole) {
	return false, model.PlayerRole_PLAYER_ROLE_UNSPECIFIED
}

func (ge *GameEngine) initializeGameStateFromScript(gameConfig loader.GameConfig, playerMap map[int32]*model.Player) {
	characters := make(map[int32]*model.Character)
	for _, charInScript := range loader.Script(gameConfig).Characters {

		charData, err := loader.Get[*model.CharacterConfig](gameConfig, charInScript.CharacterId)
		if err != nil {
			ge.logger.Warn("character config not found", zap.Int32("characterID", charInScript.CharacterId))
			continue
		}

		abilities := make([]*model.Ability, 0)

		for _, ab := range charData.AbilityIds {
			ability, err := loader.Get[*model.AbilityConfig](gameConfig, ab)
			if err != nil {
				ge.logger.Warn("ability config not found", zap.Int32("abilityID", ab))
				continue
			}
			abilities = append(abilities, &model.Ability{
				Config:           ability,
				UsedThisLoop:     false,
				OwnerCharacterId: 0,
			})
		}

		characters[charInScript.CharacterId] = &model.Character{
			Config:          charData,
			CurrentLocation: charInScript.InitialLocation,
			Paranoia:        charInScript.InitialParanoia,
			Goodwill:        charInScript.InitialGoodwill,
			Intrigue:        charInScript.InitialIntrigue,
			HiddenRole:      charInScript.HiddenRole,
			Abilities:       abilities,
			IsAlive:         true,
			InPanicMode:     false,
			Traits:          charData.Traits,
		}
	}

	ge.GameState = &model.GameState{
		GameId:                  "new_game", // Should be generated
		Characters:              characters,
		Players:                 playerMap,
		CurrentDay:              1,
		CurrentLoop:             1,
		CurrentPhase:            ge.currentPhase.Type(),
		ActiveTragedies:         make(map[int32]bool),
		PreventedTragedies:      make(map[int32]bool),
		PlayedCardsThisDay:      make(map[int32]*model.Card),
		PlayedCardsThisLoop:     make(map[int32]bool),
		LastUpdateTime:          time.Now().Unix(),
		DayEvents:               make([]*model.GameEvent, 0),
		LoopEvents:              make([]*model.GameEvent, 0),
		CharacterParanoiaLimits: make(map[int32]int32),
		CharacterGoodwillLimits: make(map[int32]int32),
		CharacterIntrigueLimits: make(map[int32]int32),
	}
}

func (ge *GameEngine) dealInitialCards() {
	script := ge.gameConfig.GetScript()
	if script == nil {
		ge.logger.Error("cannot deal cards, script not loaded")
		return
	}

	mastermind := ge.getMastermindPlayer()
	if mastermind != nil {
		for _, cardID := range script.MastermindCardIds {
			cardConfig, err := loader.Get[*model.CardConfig](ge.gameConfig, cardID)
			if err != nil {
				ge.logger.Warn("mastermind card config not found", zap.Int32("cardID", cardID))
				continue
			}
			mastermind.Hand = append(mastermind.Hand, &model.Card{Config: cardConfig})
		}
	}

	protagonists := ge.getProtagonistPlayers()
	for _, protagonist := range protagonists {
		for _, cardID := range script.ProtagonistCardIds {
			cardConfig, err := loader.Get[*model.CardConfig](ge.gameConfig, cardID)
			if err != nil {
				ge.logger.Warn("protagonist card config not found", zap.Int32("cardID", cardID))
				continue
			}
			protagonist.Hand = append(protagonist.Hand, &model.Card{Config: cardConfig})
		}
	}
}
