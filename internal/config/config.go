package config

// Config holds application configuration
type Config struct {
	AppName      string `yaml:"app_name"`
	Version      string `yaml:"version"`
	DatabasePath string `yaml:"database_path"`
	PlayerName   string `yaml:"player_name"`
	TeamName     string `yaml:"team_name"`
}

// NewDefaultConfig returns a config with default values
func NewDefaultConfig() Config {
	return Config{
		AppName:      "EU-CLAMS",
		Version:      "0.1.0",
		DatabasePath: "./data/db.yaml",
		PlayerName:   "",
	}
}
