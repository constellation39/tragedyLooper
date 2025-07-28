package loader

import (
	"fmt"
	"sync"

	v1 "tragedylooper/internal/game/proto/v1"
)

type Loader interface {
	LoadGameDataAccessor(name string) (GameConfigAccessor, error)
}

type GameConfigAccessor interface {
	GetScript() *v1.ScriptConfig

	GetAbilities() map[int32]*v1.AbilityConfig
	GetCards() map[int32]*v1.CardConfig
	GetCharacters() map[int32]*v1.CharacterConfig
	GetIncidents() map[int32]*v1.IncidentConfig
}

type cfgPtr interface {
	*v1.AbilityConfig |
		*v1.CardConfig |
		*v1.CharacterConfig |
		*v1.IncidentConfig
}

func Script(acc GameConfigAccessor) *v1.ScriptConfig {
	return acc.GetScript()
}

func Get[T cfgPtr](acc GameConfigAccessor, id int32) (T, error) {
	m, err := pickMap[T](acc)
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

func All[T cfgPtr](acc GameConfigAccessor) (map[int32]T, error) {
	return pickMap[T](acc)
}

// 根据目标类型选出正确的 map
func pickMap[T cfgPtr](acc GameConfigAccessor) (map[int32]T, error) {
	var zero T
	switch any(zero).(type) {
	case *v1.AbilityConfig:
		return any(acc.GetAbilities()).(map[int32]T), nil
	case *v1.CardConfig:
		return any(acc.GetCards()).(map[int32]T), nil
	case *v1.CharacterConfig:
		return any(acc.GetCharacters()).(map[int32]T), nil
	case *v1.IncidentConfig:
		return any(acc.GetIncidents()).(map[int32]T), nil
	default:
		return nil, fmt.Errorf("unsupported config type")
	}
}

type jsonLoader struct {
	dataDir  string
	sync.Map // map[string]*gameConfigAccessor
}

func NewJSONLoader(dir string) Loader { return &jsonLoader{dataDir: dir} }

func (l *jsonLoader) LoadGameDataAccessor(name string) (GameConfigAccessor, error) {
	// 已经加载过直接返回
	if v, ok := l.Load(name); ok {
		return v.(*gameConfigAccessor), nil
	}

	g := &gameConfigAccessor{}
	// 并发加载 5 份文件
	if err := parallel(
		taskOf(&g.abilities, func() (*v1.AbilityConfigLib, error) { return LoadAbility(l.dataDir) }),
		taskOf(&g.cards, func() (*v1.CardConfigLib, error) { return LoadCard(l.dataDir) }),
		taskOf(&g.characters, func() (*v1.CharacterConfigLib, error) { return LoadCharacter(l.dataDir) }),
		taskOf(&g.incidents, func() (*v1.IncidentConfigLib, error) { return LoadIncidents(l.dataDir) }),
		taskOf(&g.script, func() (*v1.ScriptConfig, error) { return LoadScript(l.dataDir, name) }),
	); err != nil {
		return nil, err
	}

	l.Store(name, g)
	return g, nil
}

type gameConfigAccessor struct {
	abilities  *v1.AbilityConfigLib
	cards      *v1.CardConfigLib
	characters *v1.CharacterConfigLib
	incidents  *v1.IncidentConfigLib
	script     *v1.ScriptConfig
}

func (g *gameConfigAccessor) GetAbilities() map[int32]*v1.AbilityConfig { return g.abilities.Abilities }
func (g *gameConfigAccessor) GetCards() map[int32]*v1.CardConfig        { return g.cards.Cards }
func (g *gameConfigAccessor) GetCharacters() map[int32]*v1.CharacterConfig {
	return g.characters.Characters
}
func (g *gameConfigAccessor) GetIncidents() map[int32]*v1.IncidentConfig {
	return g.incidents.Incidents
}
func (g *gameConfigAccessor) GetScript() *v1.ScriptConfig { return g.script }

type loadTask interface {
	run(chan<- error, *sync.WaitGroup)
}

type genericTask[T any] struct {
	dst **T
	fn  func() (*T, error)
}

func taskOf[T any](dst **T, fn func() (*T, error)) loadTask {
	return &genericTask[T]{dst: dst, fn: fn}
}

func (t *genericTask[T]) run(errs chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	v, err := t.fn()
	if err != nil {
		errs <- err
		return
	}
	*t.dst = v
}

func parallel(tasks ...loadTask) error {
	errs := make(chan error, len(tasks))
	var wg sync.WaitGroup
	wg.Add(len(tasks))

	for _, t := range tasks {
		go t.run(errs, &wg)
	}

	wg.Wait()
	close(errs)

	for e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
