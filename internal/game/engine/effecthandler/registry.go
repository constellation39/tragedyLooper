package effecthandler

import (
	"fmt"
	"reflect"
	model "tragedylooper/pkg/proto/v1"
)

// effectHandlers maps the type of an effect's payload to its corresponding handler.
var effectHandlers = map[reflect.Type]EffectHandler{
	reflect.TypeOf(&model.Effect_AdjustStat{}):   &AdjustStatHandler{},
	reflect.TypeOf(&model.Effect_MoveCharacter{}): &MoveCharacterHandler{},
	reflect.TypeOf(&model.Effect_AddTrait{}):      &AddTraitHandler{},
	reflect.TypeOf(&model.Effect_RemoveTrait{}):   &RemoveTraitHandler{},
	reflect.TypeOf(&model.Effect_CompoundEffect{}): &CompoundEffectHandler{},
	// Add other handlers here as they are created.
}

// GetEffectHandler returns the appropriate handler for the given effect's type.
func GetEffectHandler(effect *model.Effect) (EffectHandler, error) {
	if effect == nil || effect.EffectType == nil {
		return nil, fmt.Errorf("effect or effect type is nil")
	}
	t := reflect.TypeOf(effect.EffectType)
	handler, ok := effectHandlers[t]
	if !ok {
		return nil, fmt.Errorf("no effect handler found for type %s", t)
	}
	return handler, nil
}

// GetEffectDescription is a helper to get the description for an effect.
func GetEffectDescription(ge GameEngine, effect *model.Effect) string {
	handler, err := GetEffectHandler(effect)
	if err != nil {
		// Consider logging this error
		return "(Unknown Effect)"
	}
	return handler.GetDescription(effect)
}
