package config

// Config holds application configuration
type Config struct {
	AppName             string `yaml:"app_name"`
	DatabasePath        string `yaml:"database_path"`
	PlayerName          string `yaml:"player_name"`
	TeamName            string `yaml:"team_name"`
	EnableScreenshots   bool   `yaml:"enable_screenshots"`
	ScreenshotDirectory string `yaml:"screenshot_directory"`
	GameWindowTitle     string `yaml:"game_window_title"`
}

// NewDefaultConfig returns a config with default values
func NewDefaultConfig() Config {
	return Config{
		AppName:             "EU-CLAMS",
		DatabasePath:        "./data/db.yaml",
		PlayerName:          "",
		EnableScreenshots:   true,
		ScreenshotDirectory: "./data/screenshots",
		GameWindowTitle:     "Entropia Universe Client",
	}
}
