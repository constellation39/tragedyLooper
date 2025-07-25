package model

// GameMutator 定义了一个接口，Effect 可以通过它来修改游戏状态。
// 这避免了 model 包对 engine 包的直接依赖，从而打破了循环依赖。
type GameMutator interface {
	GetCharacter(id string) (*Character, bool)
	SetCharacterLocation(id string, location LocationType)
	AdjustCharacterParanoia(id string, amount int) (newValue int)
	AdjustCharacterGoodwill(id string, amount int) (newValue int)
	AdjustCharacterIntrigue(id string, amount int) (newValue int)
	PublishEvent(eventType EventType, payload interface{})
}
