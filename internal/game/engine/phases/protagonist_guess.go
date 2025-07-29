package phases

import (
	model "tragedylooper/internal/game/proto/v1"
)

// --- ProtagonistGuessPhase ---
type ProtagonistGuessPhase struct{ basePhase }

func (p *ProtagonistGuessPhase) Type() model.GamePhase { return model.GamePhase_PROTAGONIST_GUESS }
