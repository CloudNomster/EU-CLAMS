package config

// Config holds application configuration
type Config struct {
	AppName             string  `yaml:"app_name"`
	DatabasePath        string  `yaml:"database_path"`
	PlayerName          string  `yaml:"player_name"`
	TeamName            string  `yaml:"team_name"`
	ChatLogPath         string  `yaml:"chat_log_path"` // Path to Entropia Universe chat.log file
	EnableScreenshots   bool    `yaml:"enable_screenshots"`
	ScreenshotDirectory string  `yaml:"screenshot_directory"`
	ScreenshotDelay     float64 `yaml:"screenshot_delay"` // Delay in seconds before taking a screenshot
	GameWindowTitle     string  `yaml:"game_window_title"`
	EnableWebServer     bool    `yaml:"enable_web_server"`
	WebServerPort       int     `yaml:"web_server_port"`
}

// NewDefaultConfig returns a config with default values
func NewDefaultConfig() Config {
	return Config{
		AppName:             "EU-CLAMS",
		DatabasePath:        "./data/db.yaml",
		PlayerName:          "",
		TeamName:            "",
		ChatLogPath:         "", // Empty by default, will auto-detect if not specified
		EnableScreenshots:   true,
		ScreenshotDirectory: "./data/screenshots",
		ScreenshotDelay:     0.6, // Default to 600ms
		GameWindowTitle:     "Entropia Universe Client",
		EnableWebServer:     false,
		WebServerPort:       8080,
	}
}
