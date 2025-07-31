package phasehandler

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
		panic("phasehandler: Register of nil phasehandler")
	}
	if _, dup := phases[p.Type()]; dup {
		panic("phasehandler: Register called twice for phasehandler " + p.Type().String())
	}
	phases[p.Type()] = p
}

// GetPhase 根据类型获取一个游戏阶段的实例。
func GetPhase(phaseType model.GamePhase) Phase {
	p, ok := phases[phaseType]
	if !ok {
		panic(fmt.Sprintf("phasehandler: unknown phasehandler %s", phaseType.String()))
	}
	return p
}
