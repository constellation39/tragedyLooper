package character

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// MoveCharacter attempts to move a character, checking for restrictions and triggering a move event.
func MoveCharacter(logger *zap.Logger, triggerer EventTriggerer, gs *model.GameState, char *model.Character, dx, dy int) {
	newLoc, ok := calculateNewLocation(gs, char.CurrentLocation, dx, dy)
	if !ok {
		logger.Warn("character in unknown location", zap.String("char", char.Config.Name))
		return
	}

	if newLoc == char.CurrentLocation {
		return
	}

	// Check for movement restrictions
	for _, rule := range char.Config.Rules {
		if smr, ok := rule.Effect.(*model.CharacterRule_SpecialMovementRule); ok {
			for _, restricted := range smr.SpecialMovementRule.RestrictedLocations {
				if restricted == newLoc {
					logger.Info("character movement restricted", zap.String("char", char.Config.Name), zap.String("location", newLoc.String()))
					return // Movement forbidden
				}
			}
		}
	}

	triggerer.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_CHARACTER_MOVED, &model.EventPayload{
		Payload: &model.EventPayload_CharacterMoved{CharacterMoved: &model.CharacterMovedEvent{
			CharacterId: char.Config.Id,
			NewLocation: newLoc,
		}},
	})
}

func calculateNewLocation(gs *model.GameState, current model.LocationType, dx, dy int) (model.LocationType, bool) {
	// This is a placeholder for the actual grid logic, which should be defined centrally.
	// We assume a simple 2x2 grid for now.
	grid := map[model.LocationType]struct{ X, Y int }{
		model.LocationType_LOCATION_TYPE_SHRINE:   {0, 0},
		model.LocationType_LOCATION_TYPE_SCHOOL:   {1, 0},
		model.LocationType_LOCATION_TYPE_HOSPITAL: {0, 1},
		model.LocationType_LOCATION_TYPE_CITY:     {1, 1},
	}

	startPos, ok := grid[current]
	if !ok {
		return model.LocationType_LOCATION_TYPE_UNSPECIFIED, false
	}

	// Calculate the new position, wrapping around the 2x2 grid.
	newX := (startPos.X + dx) % 2
	newY := (startPos.Y + dy) % 2

	for loc, pos := range grid {
		if pos.X == newX && pos.Y == newY {
			return loc, true
		}
	}

	return model.LocationType_LOCATION_TYPE_UNSPECIFIED, false // Should be unreachable
}

// EventTriggerer is an interface to allow the character manager to trigger events.
type EventTriggerer interface {
	TriggerEvent(eventType model.GameEventType, payload *model.EventPayload)
}
