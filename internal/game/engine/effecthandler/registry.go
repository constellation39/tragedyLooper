package effecthandler

import (
	"fmt"
	"reflect"

	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// effectHandlers maps the type of an effect's payload to its corresponding handler.
var effectHandlers = make(map[reflect.Type]EffectHandler)

// Register is a generic function called by handlers to register themselves.
// It uses a type parameter T to infer the concrete effect payload type.
func Register[T any](handler EffectHandler) {
	var zero T
	// We use reflect.TypeOf on a zero value of type T.
	// For oneof fields like `Effect_AdjustStat`, T will be `*model.Effect_AdjustStat`.
	// reflect.TypeOf will correctly return the pointer type, which is what we use as a key.
	t := reflect.TypeOf(zero)
	if _, exists := effectHandlers[t]; exists {
		// Optional: panic or log if a handler for a type is registered more than once.
		panic(fmt.Sprintf("handler for type %v already registered", t))
	}
	effectHandlers[t] = handler
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

// withHandler is a higher-order function that retrieves the handler for an effect
// and executes a given function with it. This reduces boilerplate in helper functions.
func withHandler[T any](ge GameEngine, effect *model.Effect, fn func(EffectHandler) (T, error)) (T, error) {
	var zero T
	handler, err := GetEffectHandler(effect)
	if err != nil {
		return zero, err
	}
	return fn(handler)
}

// GetEffectDescription is a helper to get the description for an effect.
func GetEffectDescription(ge GameEngine, effect *model.Effect) string {
	description, err := withHandler(ge, effect, func(h EffectHandler) (string, error) {
		return h.GetDescription(effect), nil
	})
	if err != nil {
		// Log the error, as GetDescription is not expected to fail but the handler might be missing.
		// logger.Error("failed to get effect description", "error", err)
		return "(Unknown Effect)"
	}
	return description
}

// ResolveChoices is a helper to resolve choices for an effect.
func ResolveChoices(ge GameEngine, effect *model.Effect, ctx *EffectContext) ([]*model.Choice, error) {
	return withHandler(ge, effect, func(h EffectHandler) ([]*model.Choice, error) {
		return h.ResolveChoices(ge, effect, ctx)
	})
}

// ApplyEffect is a helper to apply an effect.
func ApplyEffect(ge GameEngine, effect *model.Effect, ctx *EffectContext) error {
	_, err := withHandler(ge, effect, func(h EffectHandler) (any, error) {
		return nil, h.Apply(ge, effect, ctx)
	})
	return err
}
