package loader

import (
	"fmt"
	v1 "tragedylooper/pkg/proto/v1"
)

// Repository holds all the game data.
type Repository struct {
	abilities  map[int32]*v1.AbilityConfig
	cards      map[int32]*v1.CardConfig
	characters map[int32]*v1.CharacterConfig
	incidents  map[int32]*v1.IncidentConfig
	script     *v1.ScriptConfig
}

func NewRepository() *Repository {
	return &Repository{
		abilities:  make(map[int32]*v1.AbilityConfig),
		cards:      make(map[int32]*v1.CardConfig),
		characters: make(map[int32]*v1.CharacterConfig),
		incidents:  make(map[int32]*v1.IncidentConfig),
	}
}

func (r *Repository) GetScript() *v1.ScriptConfig {
	return r.script
}

func (r *Repository) GetAbilities() map[int32]*v1.AbilityConfig {
	return r.abilities
}

func (r *Repository) GetCards() map[int32]*v1.CardConfig {
	return r.cards
}

func (r *Repository) GetCharacters() map[int32]*v1.CharacterConfig {
	return r.characters
}

func (r *Repository) GetIncidents() map[int32]*v1.IncidentConfig {
	return r.incidents
}

type cfgPtr interface {
	*v1.AbilityConfig |
		*v1.CardConfig |
		*v1.CharacterConfig |
		*v1.IncidentConfig
}

func Get[T cfgPtr](r GameDataAccessor, id int32) (T, error) {
	m, err := pickMap[T](r)
	if err != nil {
		var zero T
		return zero, err
	}
	v, ok := m[id]
	if !ok {
		var zero T
		return zero, fmt.Errorf("id=%d not found", id)
	}
	return v, nil
}

func All[T cfgPtr](r GameDataAccessor) (map[int32]T, error) {
	return pickMap[T](r)
}

func pickMap[T cfgPtr](r GameDataAccessor) (map[int32]T, error) {
	var zero T
	switch any(zero).(type) {
	case *v1.AbilityConfig:
		return any(r.GetAbilities()).(map[int32]T), nil
	case *v1.CardConfig:
		return any(r.GetCards()).(map[int32]T), nil
	case *v1.CharacterConfig:
		return any(r.GetCharacters()).(map[int32]T), nil
	case *v1.IncidentConfig:
		return any(r.GetIncidents()).(map[int32]T), nil
	default:
		return nil, fmt.Errorf("unsupported config type")
	}
}
