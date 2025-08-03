package schemas

import (
	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// #ConstrainedScript wraps the generated v1.#Script with validation rules.
#ConstrainedScript: v1.#Script // & {
	// Validate that all map keys fall within their designated ID ranges.
	// CUE converts proto map keys (int32) to strings, so we must convert them back for numeric comparison.

	// let _ = {
	// 	for k, _ in main_plots {
	// 		let id = strconv.Atoi(k)
	// 		if id < 1000 || id > 1999 {
	// 			_|_ // Error: Main plot ID \(k) is out of the 1000-1999 range.
	// 		}
	// 	}
	// 	for k, _ in sub_plots {
	// 		let id = strconv.Atoi(k)
	// 		if id < 2000 || id > 2999 {
	// 			_|_ // Error: Sub plot ID \(k) is out of the 2000-2999 range.
	// 		}
	// 	}
	// 	for k, _ in roles {
	// 		let id = strconv.Atoi(k)
	// 		if id < 3000 || id > 3999 {
	// 			_|_ // Error: Role ID \(k) is out of the 3000-3999 range.
	// 		}
	// 	}
	// 	for k, _ in incidents {
	// 		let id = strconv.Atoi(k)
	// 		if id < 4000 || id > 4999 {
	// 			_|_ // Error: Incident ID \(k) is out of the 4000-4999 range.
	// 		}
	// 	}
	// 	for k, _ in characters {
	// 		let id = strconv.Atoi(k)
	// 		if id < 5000 || id > 5999 {
	// 			_|_ // Error: Character ID \(k) is out of the 5000-5999 range.
	// 		}
	// 	}
	// 	for k, _ in mastermind_cards {
	// 		let id = strconv.Atoi(k)
	// 		if id < 6000 || id > 6999 {
	// 			_|_ // Error: Mastermind card ID \(k) is out of the 6000-6999 range.
	// 		}
	// 	}
	// 	for k, _ in protagonist_cards {
	// 		let id = strconv.Atoi(k)
	// 		if id < 7000 || id > 7999 {
	// 			_|_ // Error: Protagonist card ID \(k) is out of the 7000-7999 range.
	// 		}
	// 	}
	// 	for k, _ in scripts {
	// 		let id = strconv.Atoi(k)
	// 		if id < 8000 || id > 8999 {
	// 			_|_ // Error: Script config ID \(k) is out of the 8000-8999 range.
	// 		}
	// 	}
	// }
//}