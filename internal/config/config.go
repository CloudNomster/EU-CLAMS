package config

// Config holds application configuration
type Config struct {
	AppName      string `yaml:"app_name"`
	DatabasePath string `yaml:"database_path"`
	PlayerName   string `yaml:"player_name"`
	TeamName     string `yaml:"team_name"`
}

// NewDefaultConfig returns a config with default values
func NewDefaultConfig() Config {
	return Config{
		AppName:      "EU-CLAMS",
		DatabasePath: "./data/db.yaml",
		PlayerName:   "",
	}
}
