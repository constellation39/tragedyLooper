package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"tragedylooper/internal/game/proto/v1"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/gocode/gocodec"
)

func loadDataFromCue[T any](filePath string) (T, error) {
	var zero T

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return zero, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	ctx := cuecontext.New()
	val := ctx.CompileBytes(bytes, cue.Filename(filePath))
	if err := val.Err(); err != nil {
		return zero, fmt.Errorf("failed to compile cue from %s: %w", filePath, err)
	}

	var items T
	codec := &gocodec.Codec{Runtime: ctx}
	if err := codec.Decode(val, &items); err != nil {
		return zero, fmt.Errorf("failed to decode cue from %s: %w", filePath, err)
	}

	return items, nil
}

func LoadAbility(dataDir string) (*v1.AbilityLib, error) {
	filePath := filepath.Join(dataDir, "AbilityConfigLib.cue")
	abilities, err := loadDataFromCue[*v1.AbilityLib](filePath)
	if err != nil {
		return nil, err
	}
	return abilities, nil
}

func LoadCard(dataDir string) (*v1.CardLib, error) {
	filePath := filepath.Join(dataDir, "CardConfigLib.cue")
	cards, err := loadDataFromCue[*v1.CardLib](filePath)
	if err != nil {
		return nil, err
	}

	return cards, nil
}

func LoadCharacter(dataDir string) (*v1.CharacterLib, error) {
	filePath := filepath.Join(dataDir, "CharacterConfigLib.cue")
	characters, err := loadDataFromCue[*v1.CharacterLib](filePath)
	if err != nil {
		return nil, err
	}
	return characters, nil
}

func LoadScript(dataDir, scriptName string) (*v1.Script, error) {
	filePath := filepath.Join(dataDir, "ScriptConfig", scriptName+".cue")
	script, err := loadDataFromCue[*v1.Script](filePath)
	if err != nil {
		return nil, err
	}
	return script, nil
}

func LoadIncidents(dataDir string) (*v1.IncidentConfigLib, error) {
	filePath := filepath.Join(dataDir, "IncidentConfigLib.cue")
	incidents, err := loadDataFromCue[*v1.IncidentConfigLib](filePath)
	if err != nil {
		return nil, err
	}
	return incidents, nil
}
