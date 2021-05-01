package config

import (
	"encoding/json"
	"io/ioutil"
)

// ConfigLoader load config from disk
func ConfigLoader(path string) ([]*Config, error) {
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var res []*Config
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ConfigStore backup config to disk
func ConfigStore(path string, cfgs []*Config) error {
	body, err := json.Marshal(cfgs)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, body, 0644)
}
