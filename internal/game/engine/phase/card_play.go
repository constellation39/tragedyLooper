package phase

import (
	"time"
	model "tragedylooper/pkg/proto/v1"
)

// --- CardPlayPhase ---
type CardPlayPhase struct{ basePhase }

func (p *CardPlayPhase) Type() model.GamePhase { return model.GamePhase_CARD_PLAY }
func (p *CardPlayPhase) Enter(ge GameEngine) Phase {
	// Players have a certain amount of time to play their cards.
	return nil
}
func (p *CardPlayPhase) HandleTimeout(ge GameEngine) Phase {
	// If players don't act in time, we might auto-pass for them.
	return &CardRevealPhase{}
}
func (p *CardPlayPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase {
	if ge.AreAllPlayersReady() {
		return &CardRevealPhase{}
	}
	return nil
}
func (p *CardPlayPhase) TimeoutDuration() time.Duration { return 30 * time.Second } // Example timeout
