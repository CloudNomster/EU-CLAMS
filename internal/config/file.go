package config

import (
	"os"
	
	"gopkg.in/yaml.v3"
)

// LoadConfigFromFile loads configuration from a YAML file
func LoadConfigFromFile(path string) (Config, error) {
	var config Config
		data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	
	return config, nil
}

// SaveConfigToFile saves configuration to a YAML file
func (c Config) SaveConfigToFile(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}
