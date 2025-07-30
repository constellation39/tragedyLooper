package engine

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

type characterManager struct {
	engine *GameEngine
}

func newCharacterManager(engine *GameEngine) *characterManager {
	return &characterManager{
		engine: engine,
	}
}

func (cm *characterManager) MoveCharacter(char *model.Character, dx, dy int) {
	startPos, ok := LocationGrid[char.CurrentLocation]
	if !ok {
		cm.engine.logger.Warn("character in unknown location", zap.String("char", char.Config.Name))
		return
	}

	// Calculate the new position, wrapping around the 2x2 grid.
	newX := (startPos.X + dx) % 2
	newY := (startPos.Y + dy) % 2

	var newLoc model.LocationType
	for loc, pos := range LocationGrid {
		if pos.X == newX && pos.Y == newY {
			newLoc = loc
			break
		}
	}

	if newLoc != model.LocationType_LOCATION_TYPE_UNSPECIFIED && newLoc != char.CurrentLocation {
		// Check for movement restrictions
		for _, rule := range char.Config.Rules {
			if smr, ok := rule.Effect.(*model.CharacterRule_SpecialMovementRule); ok {
				for _, restricted := range smr.SpecialMovementRule.RestrictedLocations {
					if restricted == newLoc {
						cm.engine.logger.Info("character movement restricted", zap.String("char", char.Config.Name), zap.String("location", newLoc.String()))
						return // Movement forbidden
					}
				}
			}
		}

		cm.engine.ApplyAndPublishEvent(model.GameEventType_CHARACTER_MOVED, &model.EventPayload{
			Payload: &model.EventPayload_CharacterMoved{CharacterMoved: &model.CharacterMovedEvent{
				CharacterId: char.Config.Id,
				NewLocation: newLoc,
			}},
		})
	}
}

func (cm *characterManager) GetCharactersInLocation(location model.LocationType) []int32 {
	var charIDs []int32
	for id, char := range cm.engine.GameState.Characters {
		if char.CurrentLocation == location {
			charIDs = append(charIDs, id)
		}
	}
	return charIDs
}

func (cm *characterManager) GetAllCharacterIDs() []int32 {
	var charIDs []int32
	for id := range cm.engine.GameState.Characters {
		charIDs = append(charIDs, id)
	}
	return charIDs
}
