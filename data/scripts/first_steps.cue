package data

import (
	"github.com/constellation39/tragedylooper/data/schemas"
	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

first_steps: data.#ConstrainedScript & {
	name:          "First Steps"
	loop_count:    3
	days_per_loop: 4
	win_conditions: [{
		type: v1.#EndConditionType_END_CONDITION_TYPE_ALL_INCIDENTS_PREVENTED
	}]
	lose_conditions: [{
		type: v1.#EndConditionType_END_CONDITION_TYPE_SPECIFIC_INCIDENT_TRIGGERED
	}]
	mastermind_card_ids: [3, 4, 7, 8, 11, 13, 15]
	protagonist_card_ids: [1, 2, 5, 6, 9, 10, 12, 14]
	main_plot: {
		plot_type:   v1.#PlotType_PLOT_TYPE_MAIN_PLOT
		name:        "MURDER_MYSTERY"
		description: "The Killer, who is the Doctor, will murder the Target, who is the High School Girl, if they are in the same location and the Killer's paranoia is high enough."
		incident_ids: []
	}
	sub_plots: [{
		plot_type:   v1.#PlotType_PLOT_TYPE_SUB_PLOT
		name:        "HOSPITAL_CONSPIRACY"
		description: "The Hospital Director, who is the Politician, is covering up a medical error. If their intrigue reaches a certain level, they will cause a scandal."
		incident_ids: []
	}]
	characters: [{
		character_id:     1
		hidden_role:      v1.#RoleType_ROLE_TYPE_PERSON
		initial_location: v1.#LocationType_LOCATION_TYPE_SCHOOL
	}, {
		character_id:     2
		hidden_role:      v1.#RoleType_ROLE_TYPE_KILLER
		initial_location: v1.#LocationType_LOCATION_TYPE_HOSPITAL
	}, {
		character_id:     3
		hidden_role:      v1.#RoleType_ROLE_TYPE_PERSON
		initial_location: v1.#LocationType_LOCATION_TYPE_CITY
	}, {
		character_id:     35
		hidden_role:      v1.#RoleType_ROLE_TYPE_CONSPIRACY_THEORIST
		initial_location: v1.#LocationType_LOCATION_TYPE_CITY
	}, {
		character_id:     4
		hidden_role:      v1.#RoleType_ROLE_TYPE_FRIEND
		initial_location: v1.#LocationType_LOCATION_TYPE_CITY
	}, {
		character_id:     5
		hidden_role:      v1.#RoleType_ROLE_TYPE_PERSON
		initial_location: v1.#LocationType_LOCATION_TYPE_CITY
	}]
	incidents: [{
		day:                  2
		name:                 "Murder"
		culprit_character_id: 2
		victim_character_id:  1
		description:          "A murder occurs at the hospital."
	}, {
		day:                  3
		name:                 "Suicide"
		culprit_character_id: 35
		description:          "A key witness commits suicide."
	}, {
		name: "Culprit"
		trigger_conditions: [{
			stat_condition: {
				target: {
					selector_type: "SELECTOR_TYPE_SPECIFIC_CHARACTER"
					character_id:  2
				}
				stat_type:  "STAT_TYPE_PARANOIA"
				comparator: "COMPARATOR_GREATER_THAN_OR_EQUAL"
				value:      3
			}
		}]
		effect: add_trait: {
			target: {
				selector_type: "SELECTOR_TYPE_SPECIFIC_CHARACTER"
				character_id:  2
			}
			trait: "Culprit"
		}
	}]
}
