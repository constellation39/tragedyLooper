package engine

import (
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

type CharacterManager struct {
	engine *GameEngine
}

func NewCharacterManager(engine *GameEngine) *CharacterManager {
	return &CharacterManager{
		engine: engine,
	}
}

func (cm *CharacterManager) MoveCharacter(char *model.Character, dx, dy int) {
	startPos, ok := LocationGrid[char.CurrentLocation]
	if !ok {
		cm.engine.logger.Warn("character in unknown location", zap.String("char", char.Config.Name))
		return
	}

	// 计算新位置，在 2x2 网格上环绕。
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
		// 检查移动限制
		for _, rule := range char.Config.Rules {
			if smr, ok := rule.Effect.(*model.CharacterRule_SpecialMovementRule); ok {
				for _, restricted := range smr.SpecialMovementRule.RestrictedLocations {
					if restricted == newLoc {
						cm.engine.logger.Info("character movement restricted", zap.String("char", char.Config.Name), zap.String("location", newLoc.String()))
						return // 禁止移动
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

func (cm *CharacterManager) GetCharactersInLocation(location model.LocationType) []int32 {
	var charIDs []int32
	for id, char := range cm.engine.GameState.Characters {
		if char.CurrentLocation == location {
			charIDs = append(charIDs, id)
		}
	}
	return charIDs
}

func (cm *CharacterManager) GetAllCharacterIDs() []int32 {
	var charIDs []int32
	for id := range cm.engine.GameState.Characters {
		charIDs = append(charIDs, id)
	}
	return charIDs
}
