package config

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

func (c *Config) FromJSON(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("Failed to parse JSON: %s", err.Error())
	}

	return nil
}

func (c *Config) FromYAML(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("Failed to parse YAML: %s", err.Error())
	}

	return nil
}

func (c *Config) ToJSON() ([]byte, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert to JSON: %s", err.Error())
	}

	return data, nil
}

func (c *Config) ToJSONWithIndent() ([]byte, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("Failed to convert to JSON: %s", err.Error())
	}

	return data, nil
}

func (c *Config) ToYAML() ([]byte, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert to YAML: %s", err.Error())
	}

	return data, nil
}
