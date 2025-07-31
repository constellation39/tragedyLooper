package phase

import (
	"fmt"
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

var (
	phases = make(map[model.GamePhase]Phase)
)

// RegisterPhase 注册一个游戏阶段。
func RegisterPhase(p Phase) {
	if p == nil {
		panic("phase: Register of nil phase")
	}
	if _, dup := phases[p.Type()]; dup {
		panic("phase: Register called twice for phase " + p.Type().String())
	}
	phases[p.Type()] = p
}

// GetPhase 根据类型获取一个游戏阶段的实例。
func GetPhase(phaseType model.GamePhase) Phase {
	p, ok := phases[phaseType]
	if !ok {
		panic(fmt.Sprintf("phase: unknown phase %s", phaseType.String()))
	}
	return p
}
