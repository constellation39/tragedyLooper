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
	newLoc, ok := cm.calculateNewLocation(char.CurrentLocation, dx, dy)
	if !ok {
		cm.engine.logger.Warn("character in unknown location", zap.String("char", char.Config.Name))
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
					cm.engine.logger.Info("character movement restricted", zap.String("char", char.Config.Name), zap.String("location", newLoc.String()))
					return // Movement forbidden
				}
			}
		}
	}

	cm.engine.ApplyAndPublishEvent(model.GameEventType_GAME_EVENT_TYPE_CHARACTER_MOVED, &model.EventPayload{
		Payload: &model.EventPayload_CharacterMoved{CharacterMoved: &model.CharacterMovedEvent{
			CharacterId: char.Config.Id,
			NewLocation: newLoc,
		}},
	})

}

func (cm *characterManager) calculateNewLocation(current model.LocationType, dx, dy int) (model.LocationType, bool) {
	startPos, ok := LocationGrid[current]
	if !ok {
		return model.LocationType_LOCATION_TYPE_UNSPECIFIED, false
	}

	// Calculate the new position, wrapping around the 2x2 grid.
	newX := (startPos.X + dx) % 2
	newY := (startPos.Y + dy) % 2

	for loc, pos := range LocationGrid {
		if pos.X == newX && pos.Y == newY {
			return loc, true
		}
	}

	return model.LocationType_LOCATION_TYPE_UNSPECIFIED, false // Should be unreachable
}

func (cm *characterManager) GetCharactersInLocation(location model.LocationType) []int32 {
	charIDs := make([]int32, 0, len(cm.engine.GameState.Characters))
	for id, char := range cm.engine.GameState.Characters {
		if char.CurrentLocation == location {
			charIDs = append(charIDs, id)
		}
	}
	return charIDs
}

func (cm *characterManager) GetAllCharacterIDs() []int32 {
	charIDs := make([]int32, 0, len(cm.engine.GameState.Characters))
	for id := range cm.engine.GameState.Characters {
		charIDs = append(charIDs, id)
	}
	return charIDs
}
