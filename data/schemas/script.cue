// #ConstrainedScript wraps the generated v1.#Script with validation rules.
package schemas

import (
	"strconv"
	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// #ConstrainedScript wraps the generated v1.#Script with validation rules.
#ConstrainedScript: v1.#Script & {
	// Validate that all map keys fall within their designated ID ranges.
	// CUE converts proto map keys (int32) to strings, so we must convert them back for numeric comparison.
	main_plots: {
		[string]: v1.#MainPlot
		for k, _ in main_plots {
			let id = strconv.ParseInt(k, 10)
			if id < 1000 || id > 1999 {
				_|_ // Error: Main plot ID \(k) is out of the 1000-1999 range.
			}
		}
	}
	sub_plots: {
		[string]: v1.#SubPlot
		for k, _ in sub_plots {
			let id = strconv.ParseInt(k, 10)
			if id < 2000 || id > 2999 {
				_|_ // Error: Sub plot ID \(k) is out of the 2000-2999 range.
			}
		}
	}
	roles: {
		[string]: v1.#Role
		for k, _ in roles {
			let id = strconv.ParseInt(k, 10)
			if id < 3000 || id > 3999 {
				_|_ // Error: Role ID \(k) is out of the 3000-3999 range.
			}
		}
	}
	incidents: {
		[string]: v1.#Incident
		for k, _ in incidents {
			let id = strconv.ParseInt(k, 10)
			if id < 4000 || id > 4999 {
				_|_ // Error: Incident ID \(k) is out of the 4000-4999 range.
			}
		}
	}
	characters: {
		[string]: v1.#Character
		for k, _ in characters {
			let id = strconv.ParseInt(k, 10)
			if id < 5000 || id > 5999 {
				_|_ // Error: Character ID \(k) is out of the 5000-5999 range.
			}
		}
	}
	mastermind_cards: {
		[string]: v1.#Card
		for k, _ in mastermind_cards {
			let id = strconv.ParseInt(k, 10)
			if id < 6000 || id > 6999 {
				_|_ // Error: Mastermind card ID \(k) is out of the 6000-6999 range.
			}
		}
	}
	protagonist_cards: {
		[string]: v1.#Card
		for k, _ in protagonist_cards {
			let id = strconv.ParseInt(k, 10)
			if id < 7000 || id > 7999 {
				_|_ // Error: Protagonist card ID \(k) is out of the 7000-7999 range.
			}
		}
	}
	scripts: {
		[string]: v1.#ScriptConfig
		for k, _ in scripts {
			let id = strconv.ParseInt(k, 10)
			if id < 8000 || id > 8999 {
				_|_ // Error: Script config ID \(k) is out of the 8000-8999 range.
			}
		}
	}
}