package config

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CommandLineOverrides holds command-line flag overrides
type CommandLineOverrides struct {
	ConfigPath          *string
	PlayerName          *string
	TeamName            *string
	EnableScreenshots   *bool
	ScreenshotDirectory *string
	GameWindowTitle     *string
	EnableWebServer     *bool
	WebServerPort       *int
}

// Manager handles configuration loading, saving, and reloading
type Manager struct {
	config     Config
	configPath string
	overrides  CommandLineOverrides
	fileHash   string
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		config:    NewDefaultConfig(),
		overrides: CommandLineOverrides{},
	}
}

// LoadConfig loads configuration with command-line overrides
func (m *Manager) LoadConfig(configPath string, overrides CommandLineOverrides) (Config, error) {
	m.overrides = overrides
	m.configPath = configPath

	// If config path is provided, try to load from file
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			fileConfig, err := LoadConfigFromFile(configPath)
			if err != nil {
				return Config{}, fmt.Errorf("failed to load config from %s: %w", configPath, err)
			}
			m.config = fileConfig
		} else {
			// File doesn't exist, use defaults
			m.config = NewDefaultConfig()
		}
	} else {
		// Check for default config.yaml
		defaultPath := "config.yaml"
		if _, err := os.Stat(defaultPath); err == nil {
			m.configPath = defaultPath
			fileConfig, err := LoadConfigFromFile(defaultPath)
			if err != nil {
				return Config{}, fmt.Errorf("failed to load config from %s: %w", defaultPath, err)
			}
			m.config = fileConfig
		} else {
			// No config file found, use defaults
			m.config = NewDefaultConfig()
			m.configPath = "config.yaml" // Default path for saving
		}
	}

	// Apply command-line overrides
	m.applyOverrides()

	// Ensure database path is absolute
	if !filepath.IsAbs(m.config.DatabasePath) {
		absPath, err := filepath.Abs(m.config.DatabasePath)
		if err == nil {
			m.config.DatabasePath = absPath
		}
	}

	// Update file hash for change detection
	m.updateFileHash()

	return m.config, nil
}

// applyOverrides applies command-line overrides to the config
func (m *Manager) applyOverrides() {
	if m.overrides.PlayerName != nil && *m.overrides.PlayerName != "" {
		m.config.PlayerName = *m.overrides.PlayerName
	}
	if m.overrides.TeamName != nil && *m.overrides.TeamName != "" {
		m.config.TeamName = *m.overrides.TeamName
	}
	if m.overrides.EnableScreenshots != nil {
		m.config.EnableScreenshots = *m.overrides.EnableScreenshots
	}
	if m.overrides.ScreenshotDirectory != nil && *m.overrides.ScreenshotDirectory != "" {
		m.config.ScreenshotDirectory = *m.overrides.ScreenshotDirectory
	}
	if m.overrides.GameWindowTitle != nil && *m.overrides.GameWindowTitle != "" {
		m.config.GameWindowTitle = *m.overrides.GameWindowTitle
	}
	if m.overrides.EnableWebServer != nil {
		m.config.EnableWebServer = *m.overrides.EnableWebServer
	}
	if m.overrides.WebServerPort != nil && *m.overrides.WebServerPort != 0 {
		m.config.WebServerPort = *m.overrides.WebServerPort
	}
}

// ReloadConfig reloads the configuration from file while preserving command-line overrides
func (m *Manager) ReloadConfig() (Config, bool, error) {
	if m.configPath == "" {
		return m.config, false, nil
	}

	// Check if file exists and get its hash
	newHash := m.calculateFileHash(m.configPath)
	if newHash == m.fileHash {
		// No change detected
		return m.config, false, nil
	}

	// Load fresh config from file
	if _, err := os.Stat(m.configPath); err != nil {
		// File no longer exists, keep current config
		return m.config, false, nil
	}

	fileConfig, err := LoadConfigFromFile(m.configPath)
	if err != nil {
		return m.config, false, fmt.Errorf("failed to reload config: %w", err)
	}

	// Update the base config but preserve command-line overrides
	m.config = fileConfig
	m.applyOverrides()

	// Ensure database path is absolute
	if !filepath.IsAbs(m.config.DatabasePath) {
		absPath, err := filepath.Abs(m.config.DatabasePath)
		if err == nil {
			m.config.DatabasePath = absPath
		}
	}

	// Update file hash
	m.fileHash = newHash

	return m.config, true, nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() Config {
	return m.config
}

// GetConfigPath returns the current config file path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// SetConfigPath sets the config file path and updates the file hash
func (m *Manager) SetConfigPath(path string) {
	m.configPath = path
	m.updateFileHash()
}

// SaveConfig saves the current configuration to file
func (m *Manager) SaveConfig() error {
	if m.configPath == "" {
		return fmt.Errorf("no config path set")
	}

	err := m.config.SaveConfigToFile(m.configPath)
	if err != nil {
		return err
	}

	// Update file hash after successful save
	m.updateFileHash()
	return nil
}

// UpdateConfig updates the configuration and optionally saves to file
func (m *Manager) UpdateConfig(newConfig Config, save bool) error {
	m.config = newConfig

	if save {
		return m.SaveConfig()
	}

	return nil
}

// HasFileChanged checks if the config file has changed since last load
func (m *Manager) HasFileChanged() bool {
	if m.configPath == "" {
		return false
	}

	currentHash := m.calculateFileHash(m.configPath)
	return currentHash != m.fileHash
}

// updateFileHash updates the stored file hash
func (m *Manager) updateFileHash() {
	m.fileHash = m.calculateFileHash(m.configPath)
}

// calculateFileHash calculates the MD5 hash of the config file
func (m *Manager) calculateFileHash(path string) string {
	if path == "" {
		return ""
	}

	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}
