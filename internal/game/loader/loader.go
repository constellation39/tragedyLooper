
package loader

import (
	"encoding/json"
	"io/ioutil"

	model "tragedylooper/internal/game/proto/v1"
)

// LoadScript loads a script from a JSON file.
func LoadScript(path string) (*model.Script, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var script model.Script
	if err := json.Unmarshal(data, &script);
 err != nil {
		return nil, err
	}

	return &script, nil
}

// LoadCharacter loads a character from a JSON file.
func LoadCharacter(path string) (*model.Character, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var character model.Character
	if err := json.Unmarshal(data, &character);
 err != nil {
		return nil, err
	}

	return &character, nil
}

// LoadAbility loads an ability from a JSON file.
func LoadAbility(path string) (*model.Ability, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ability model.Ability
	if err := json.Unmarshal(data, &ability);
 err != nil {
		return nil, err
	}

	return &ability, nil
}
