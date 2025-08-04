package condition

import (
	"testing"

	"github.com/constellation39/tragedyLooper/internal/game/engine/target"
	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
	"github.com/stretchr/testify/assert"
)

// setupTest creates a standard Checker and GameState for testing.
func setupTest() (*Checker, *v1.GameState) {
	resolver := target.NewResolver()
	checker := NewChecker(resolver)

	gs := &v1.GameState{
		Characters: map[int32]*v1.Character{
			1: {
				Id:       1,
				Name:     "Protagonist",
				Location: v1.LocationType_LOCATION_TYPE_SCHOOL,
				Paranoia: 5,
				Intrigue: 2,
				Traits:   []string{"Kind"},
				Role:     &v1.Role{Id: 101, Goodwill: 10},
			},
			2: {
				Id:       2,
				Name:     "Friend",
				Location: v1.LocationType_LOCATION_TYPE_SCHOOL,
				Paranoia: 2,
				Intrigue: 8,
				Traits:   []string{"Smart"},
				Role:     &v1.Role{Id: 102, Goodwill: 5},
			},
			3: {
				Id:       3,
				Name:     "Mystery Man",
				Location: v1.LocationType_LOCATION_TYPE_HOSPITAL,
				Paranoia: 9,
				Intrigue: 9,
				Traits:   []string{"Suspicious"},
				Role:     &v1.Role{Id: 201, Goodwill: 0},
			},
		},
		Day:   3,
		Phase: v1.GamePhase_GAME_PHASE_MAIN,
	}
	return checker, gs
}

func TestCheckStatCondition(t *testing.T) {
	checker, gs := setupTest()

	tests := []struct {
		name      string
		condition *v1.Condition
		expected  bool
	}{
		{
			name: "Paranoia greater than 3",
			condition: &v1.Condition{ConditionType: &v1.Condition_StatCondition{StatCondition: &v1.StatCondition{
				Target:     &v1.TargetSelector{Selector: &v1.TargetSelector_SpecificCharacter{SpecificCharacter: 1}},
				StatType:   v1.StatType_STAT_TYPE_PARANOIA,
				Comparator: v1.Comparator_GREATER_THAN,
				Value:      3,
			}}},
			expected: true,
		},
		{
			name: "Intrigue equal to 8",
			condition: &v1.Condition{ConditionType: &v1.Condition_StatCondition{StatCondition: &v1.StatCondition{
				Target:     &v1.TargetSelector{Selector: &v1.TargetSelector_SpecificCharacter{SpecificCharacter: 2}},
				StatType:   v1.StatType_STAT_TYPE_INTRIGUE,
				Comparator: v1.Comparator_EQUAL_TO,
				Value:      8,
			}}},
			expected: true,
		},
		{
			name: "Goodwill less than 1",
			condition: &v1.Condition{ConditionType: &v1.Condition_StatCondition{StatCondition: &v1.StatCondition{
				Target:     &v1.TargetSelector{Selector: &v1.TargetSelector_SpecificCharacter{SpecificCharacter: 3}},
				StatType:   v1.StatType_STAT_TYPE_GOODWILL,
				Comparator: v1.Comparator_LESS_THAN,
				Value:      1,
			}}},
			expected: true,
		},
		{
			name: "Protagonist paranoia > Friend paranoia",
			condition: &v1.Condition{ConditionType: &v1.Condition_StatCondition{StatCondition: &v1.StatCondition{
				Target:          &v1.TargetSelector{Selector: &v1.TargetSelector_SpecificCharacter{SpecificCharacter: 1}},
				StatType:        v1.StatType_STAT_TYPE_PARANOIA,
				Comparator:      v1.Comparator_GREATER_THAN,
				TargetToCompare: &v1.TargetSelector{Selector: &v1.TargetSelector_SpecificCharacter{SpecificCharacter: 2}},
			}}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := checker.Check(gs, tt.condition)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckLocationCondition(t *testing.T) {
	checker, gs := setupTest()

	tests := []struct {
		name      string
		condition *v1.Condition
		expected  bool
	}{
		{
			name: "Mystery Man is at Hospital",
			condition: &v1.Condition{ConditionType: &v1.Condition_LocationCondition{LocationCondition: &v1.LocationCondition{
				Target:   &v1.TargetSelector{Selector: &v1.TargetSelector_SpecificCharacter{SpecificCharacter: 3}},
				Location: v1.LocationType_LOCATION_TYPE_HOSPITAL,
			}}},
			expected: true,
		},
		{
			name: "Protagonist is at School",
			condition: &v1.Condition{ConditionType: &v1.Condition_LocationCondition{LocationCondition: &v1.LocationCondition{
				Target:   &v1.TargetSelector{Selector: &v1.TargetSelector_SpecificCharacter{SpecificCharacter: 1}},
				Location: v1.LocationType_LOCATION_TYPE_SCHOOL,
			}}},
			expected: true,
		},
		{
			name: "Friend is not at Shrine",
			condition: &v1.Condition{ConditionType: &v1.Condition_CompoundCondition{CompoundCondition: &v1.CompoundCondition{
				Operator: v1.CompoundCondition_OPERATOR_NOT,
				SubConditions: []*v1.Condition{
					{ConditionType: &v1.Condition_LocationCondition{LocationCondition: &v1.LocationCondition{
						Target:   &v1.TargetSelector{Selector: &v1.TargetSelector_SpecificCharacter{SpecificCharacter: 2}},
						Location: v1.LocationType_LOCATION_TYPE_SHRINE,
					}}},
				},
			}}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := checker.Check(gs, tt.condition)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckLocationCharacterCountCondition(t *testing.T) {
	checker, gs := setupTest()

	tests := []struct {
		name      string
		condition *v1.Condition
		expected  bool
	}{
		{
			name: "Exactly 2 characters at School",
			condition: &v1.Condition{ConditionType: &v1.Condition_LocationCharacterCountCondition{LocationCharacterCountCondition: &v1.LocationCharacterCountCondition{
				Location:   v1.LocationType_LOCATION_TYPE_SCHOOL,
				Comparator: v1.Comparator_EQUAL_TO,
				Count:      2,
			}}},
			expected: true,
		},
		{
			name: "More than 0 characters at Hospital",
			condition: &v1.Condition{ConditionType: &v1.Condition_LocationCharacterCountCondition{LocationCharacterCountCondition: &v1.LocationCharacterCountCondition{
				Location:   v1.LocationType_LOCATION_TYPE_HOSPITAL,
				Comparator: v1.Comparator_GREATER_THAN,
				Count:      0,
			}}},
			expected: true,
		},
		{
			name: "Less than 2 characters at Hospital (Mystery Man is alone)",
			condition: &v1.Condition{ConditionType: &v1.Condition_LocationCharacterCountCondition{LocationCharacterCountCondition: &v1.LocationCharacterCountCondition{
				Location:   v1.LocationType_LOCATION_TYPE_HOSPITAL,
				Comparator: v1.Comparator_LESS_THAN,
				Count:      2,
			}}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := checker.Check(gs, tt.condition)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckCompoundCondition(t *testing.T) {
	checker, gs := setupTest()

	// Condition A: Protagonist has paranoia > 4 (True)
	condA := &v1.Condition{ConditionType: &v1.Condition_StatCondition{StatCondition: &v1.StatCondition{
		Target:     &v1.TargetSelector{Selector: &v1.TargetSelector_SpecificCharacter{SpecificCharacter: 1}},
		StatType:   v1.StatType_STAT_TYPE_PARANOIA,
		Comparator: v1.Comparator_GREATER_THAN,
		Value:      4,
	}}}

	// Condition B: Friend has trait "Suspicious" (False)
	condB := &v1.Condition{ConditionType: &v1.Condition_TraitCondition{TraitCondition: &v1.TraitCondition{
		Target: &v1.TargetSelector{Selector: &v1.TargetSelector_SpecificCharacter{SpecificCharacter: 2}},
		Trait:  "Suspicious",
	}}}

	tests := []struct {
		name      string
		condition *v1.Condition
		expected  bool
	}{
		{
			name: "A AND B (True AND False)",
			condition: &v1.Condition{ConditionType: &v1.Condition_CompoundCondition{CompoundCondition: &v1.CompoundCondition{
				Operator:      v1.CompoundCondition_OPERATOR_AND,
				SubConditions: []*v1.Condition{condA, condB},
			}}},
			expected: false,
		},
		{
			name: "A OR B (True OR False)",
			condition: &v1.Condition{ConditionType: &v1.Condition_CompoundCondition{CompoundCondition: &v1.CompoundCondition{
				Operator:      v1.CompoundCondition_OPERATOR_OR,
				SubConditions: []*v1.Condition{condA, condB},
			}}},
			expected: true,
		},
		{
			name: "NOT B (NOT False)",
			condition: &v1.Condition{ConditionType: &v1.Condition_CompoundCondition{CompoundCondition: &v1.CompoundCondition{
				Operator:      v1.CompoundCondition_OPERATOR_NOT,
				SubConditions: []*v1.Condition{condB},
			}}},
			expected: true,
		},
		{
			name: "(A AND (NOT B))",
			condition: &v1.Condition{ConditionType: &v1.Condition_CompoundCondition{CompoundCondition: &v1.CompoundCondition{
				Operator: v1.CompoundCondition_OPERATOR_AND,
				SubConditions: []*v1.Condition{
					condA,
					{ConditionType: &v1.Condition_CompoundCondition{CompoundCondition: &v1.CompoundCondition{
						Operator:      v1.CompoundCondition_OPERATOR_NOT,
						SubConditions: []*v1.Condition{condB},
					}}},
				},
			}}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := checker.Check(gs, tt.condition)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
