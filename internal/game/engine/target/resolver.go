package target

import (
	"fmt"

	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

func ResolveCharacters(gs *v1.GameState, selector *v1.TargetSelector) ([]*v1.Character, error) {
	if selector == nil {
		return nil, fmt.Errorf("target selector is nil")
	}

	switch s := selector.Selector.(type) {
	case *v1.TargetSelector_SpecificCharacter:
		char, ok := gs.Characters[s.SpecificCharacter]
		if !ok {
			return nil, fmt.Errorf("character with id %d not found", s.SpecificCharacter)
		}
		return []*v1.Character{char}, nil

	case *v1.TargetSelector_CharacterWithRoleId:
		var matched []*v1.Character
		for _, char := range gs.Characters {
			if char.Config.Id == s.CharacterWithRoleId {
				matched = append(matched, char)
			}
		}
		return matched, nil

	case *v1.TargetSelector_AllCharactersAtLocation:
		var matched []*v1.Character
		for _, char := range gs.Characters {
			if char.CurrentLocation == s.AllCharactersAtLocation {
				matched = append(matched, char)
			}
		}
		return matched, nil

	case *v1.TargetSelector_AllCharacters:
		var all []*v1.Character
		for _, char := range gs.Characters {
			all = append(all, char)
		}
		return all, nil

	// --- Placeholders for event-based targets ---
	// These require an Event context to be passed into the resolver.
	case *v1.TargetSelector_TriggeringCharacter:
		return nil, fmt.Errorf("resolving triggering_character not yet implemented")
	case *v1.TargetSelector_Culprit:
		return nil, fmt.Errorf("resolving culprit not yet implemented")
	case *v1.TargetSelector_Victim:
		return nil, fmt.Errorf("resolving victim not yet implemented")
	case *v1.TargetSelector_ActionUser:
		return nil, fmt.Errorf("resolving action_user not yet implemented")
	case *v1.TargetSelector_ActionTarget:
		return nil, fmt.Errorf("resolving action_target not yet implemented")

	default:
		return nil, fmt.Errorf("unhandled target selector type: %T", s)
	}
}
