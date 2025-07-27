package data

import model "tragedylooper/internal/game/proto/v1"

var MastermindCards = []*model.Card{
	{
		Id:        3,
		Name:      "Move",
		CardType:  model.CardType_CARD_TYPE_MOVEMENT,
		OwnerRole: model.PlayerRole_PLAYER_ROLE_MASTERMIND,
		Effect: &model.Effect{
			EffectOneof: &model.Effect_MoveCharacterEffect{
				MoveCharacterEffect: &model.MoveCharacterEffect{},
			},
		},
		OncePerLoop: false,
	},
	{
		Id:        4,
		Name:      "Add Paranoia",
		CardType:  model.CardType_CARD_TYPE_PARANOIA,
		OwnerRole: model.PlayerRole_PLAYER_ROLE_MASTERMIND,
		Effect: &model.Effect{
			EffectOneof: &model.Effect_AdjustParanoiaEffect{
				AdjustParanoiaEffect: &model.AdjustParanoiaEffect{
					Amount: 1,
				},
			},
		},
		OncePerLoop: false,
	},
}

var ProtagonistCards = []*model.Card{
	{
		Id:        1,
		Name:      "Move",
		CardType:  model.CardType_CARD_TYPE_MOVEMENT,
		OwnerRole: model.PlayerRole_PLAYER_ROLE_PROTAGONIST,
		Effect: &model.Effect{
			EffectOneof: &model.Effect_MoveCharacterEffect{
				MoveCharacterEffect: &model.MoveCharacterEffect{},
			},
		},
		OncePerLoop: false,
	},
	{
		Id:        2,
		Name:      "Add Paranoia",
		CardType:  model.CardType_CARD_TYPE_PARANOIA,
		OwnerRole: model.PlayerRole_PLAYER_ROLE_PROTAGONIST,
		Effect: &model.Effect{
			EffectOneof: &model.Effect_AdjustParanoiaEffect{
				AdjustParanoiaEffect: &model.AdjustParanoiaEffect{
					Amount: 1,
				},
			},
		},
		OncePerLoop: false,
	},
}
