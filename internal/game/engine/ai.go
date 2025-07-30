package engine

import (
	"context"
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// ActionGenerator defines the interface for an AI to generate a player action.
type ActionGenerator interface {
	GenerateAction(ctx context.Context, data *ActionGeneratorContext) (*model.PlayerActionPayload, error)
}

// ActionGeneratorContext provides all necessary information for an AI to make a decision.
type ActionGeneratorContext struct {
	Player        *model.Player
	PlayerView    *model.PlayerView
	Script        *model.ScriptConfig
	AllCharacters map[int32]*model.Character
}
