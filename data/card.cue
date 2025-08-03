import (
	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)
cards: {
	"1": {
		id:         1
		name:       "Move"
		type:       v1.#CardType_CARD_TYPE_FORBID_MOVEMENT
		owner_role: v1.#PlayerRole_PLAYER_ROLE_PROTAGONIST
		effect: sub_effects: [{
			move_character: {}
		}]
		once_per_loop: false
	}
	"2": {
		id:         2
		name:       "Add Paranoia"
		type:       v1.#CardType_CARD_TYPE_PARANOIA_MINUS
		owner_role: v1.#PlayerRole_PLAYER_ROLE_PROTAGONIST
		effect: sub_effects: [{
			adjust_stat: {
				stat_type: "STAT_TYPE_PARANOIA"
				amount:    -1
			}
		}]
		once_per_loop: false
	}
	"3": {
		id:         3
		name:       "Move"
		type:       v1.#CardType_CARD_TYPE_FORBID_MOVEMENT
		owner_role: v1.#PlayerRole_PLAYER_ROLE_MASTERMIND
		effect: sub_effects: [{
			move_character: {}
		}]
		once_per_loop: false
	}
	"4": {
		id:         4
		name:       "Add Paranoia"
		type:       v1.#CardType_CARD_TYPE_PARANOIA_PLUS
		owner_role: v1.#PlayerRole_PLAYER_ROLE_MASTERMIND
		effect: sub_effects: [{
			adjust_stat: {
				target: selector_type: "SELECTOR_TYPE_ABILITY_TARGET"
				stat_type: "STAT_TYPE_PARANOIA"
				amount:    1
			}
		}]
		once_per_loop: false
	}
	"5": {
		id:         5
		name:       "Add Goodwill"
		type:       v1.#CardType_CARD_TYPE_GOODWILL_PLUS
		owner_role: v1.#PlayerRole_PLAYER_ROLE_PROTAGONIST
		effect: sub_effects: [{
			adjust_stat: {
				stat_type: "STAT_TYPE_GOODWILL"
				amount:    1
			}
		}]
		once_per_loop: false
	}
	"6": {
		id:         6
		name:       "Add Intrigue"
		type:       v1.#CardType_CARD_TYPE_INTRIGUE_PLUS
		owner_role: v1.#PlayerRole_PLAYER_ROLE_PROTAGONIST
		effect: sub_effects: [{
			adjust_stat: {
				stat_type: "STAT_TYPE_INTRIGUE"
				amount:    1
			}
		}]
		once_per_loop: false
	}
	"7": {
		id:         7
		name:       "Add Goodwill"
		type:       v1.#CardType_CARD_TYPE_GOODWILL_PLUS
		owner_role: v1.#PlayerRole_PLAYER_ROLE_MASTERMIND
		effect: sub_effects: [{
			adjust_stat: {
				stat_type: "STAT_TYPE_GOODWILL"
				amount:    1
			}
		}]
		once_per_loop: false
	}
	"8": {
		id:         8
		name:       "Add Intrigue"
		type:       v1.#CardType_CARD_TYPE_INTRIGUE_PLUS
		owner_role: v1.#PlayerRole_PLAYER_ROLE_MASTERMIND
		effect: sub_effects: [{
			adjust_stat: {
				stat_type: "STAT_TYPE_INTRIGUE"
				amount:    1
			}
		}]
		once_per_loop: false
	}
	"9": {
		id:         9
		name:       "Forbid MOVEMENT"
		type:       v1.#CardType_CARD_TYPE_FORBID_MOVEMENT
		owner_role: v1.#PlayerRole_PLAYER_ROLE_PROTAGONIST
		effect: sub_effects: [{
			forbid: forbid_type: v1.#ForbidType_FORBID_TYPE_MOVEMENT
		}]
		once_per_loop: false
	}
	"10": {
		id:         10
		name:       "Forbid Paranoia"
		type:       v1.#CardType_CARD_TYPE_FORBID_PARANOIA_INCREASE
		owner_role: v1.#PlayerRole_PLAYER_ROLE_PROTAGONIST
		effect: sub_effects: [{
			forbid: forbid_type: v1.#ForbidType_FORBID_TYPE_PARANOIA_CHANGE
		}]
		once_per_loop: false
	}
	"11": {
		id:         11
		name:       "Forbid Goodwill"
		type:       v1.#CardType_CARD_TYPE_FORBID_GOODWILL_INCREASE
		owner_role: v1.#PlayerRole_PLAYER_ROLE_MASTERMIND
		effect: sub_effects: [{
			forbid: forbid_type: v1.#ForbidType_FORBID_TYPE_GOODWILL_CHANGE
		}]
		once_per_loop: false
	}
	"12": {
		id:         12
		name:       "Forbid Intrigue"
		type:       v1.#CardType_CARD_TYPE_FORBID_INTRIGUE_INCREASE
		owner_role: v1.#PlayerRole_PLAYER_ROLE_PROTAGONIST
		effect: sub_effects: [{
			forbid: forbid_type: v1.#ForbidType_FORBID_TYPE_INTRIGUE_CHANGE
		}]
		once_per_loop: false
	}
	"13": {
		id:         13
		name:       "Forbid Paranoia"
		type:       v1.#CardType_CARD_TYPE_FORBID_PARANOIA_INCREASE
		owner_role: v1.#PlayerRole_PLAYER_ROLE_MASTERMIND
		effect: sub_effects: [{
			forbid: forbid_type: v1.#ForbidType_FORBID_TYPE_PARANOIA_CHANGE
		}]
		once_per_loop: false
	}
	"14": {
		id:         14
		name:       "Add Goodwill +2"
		type:       v1.#CardType_CARD_TYPE_GOODWILL_PLUS
		owner_role: v1.#PlayerRole_PLAYER_ROLE_PROTAGONIST
		effect: sub_effects: [{
			adjust_stat: {
				stat_type: "STAT_TYPE_GOODWILL"
				amount:    2
			}
		}]
		once_per_loop: true
	}
	"15": {
		id:         15
		name:       "Add Intrigue +2"
		type:       v1.#CardType_CARD_TYPE_INTRIGUE_PLUS
		owner_role: v1.#PlayerRole_PLAYER_ROLE_MASTERMIND
		effect: sub_effects: [{
			adjust_stat: {
				stat_type: "STAT_TYPE_INTRIGUE"
				amount:    2
			}
		}]
		once_per_loop: true
	}
}
