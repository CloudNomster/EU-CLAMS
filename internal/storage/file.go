package storage

import (
	"eu-clams/internal/logger"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SaveDatabase saves the database to a YAML file
func (db *EntropyDB) SaveDatabase(path string, logger *logger.Logger) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal the data to YAML
	data, err := yaml.Marshal(db)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write the file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	// Verify the file was written
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("failed to verify file was written: %w", err)
	}

	return nil
}

// LoadDatabase loads the database from a YAML file
func LoadDatabase(path string, logger *logger.Logger) (*EntropyDB, error) {
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist, return empty database
		if logger != nil {
			logger.Info("Database file does not exist at: %s. Creating new database.", path)
		}
		return NewEntropyDB("", ""), nil
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal the data
	var db EntropyDB
	if err := yaml.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	if logger != nil {
		logger.Info("Loaded database from: %s (%d globals)", path, len(db.Globals))
	}

	return &db, nil
}

// MergeDatabase merges a new database into this one, avoiding duplicates
func (db *EntropyDB) MergeDatabase(other *EntropyDB) int {
	if other == nil {
		return 0
	}

	added := 0
	// Use a map to track existing entries by their raw message to avoid duplicates
	existing := make(map[string]bool)
	for _, entry := range db.Globals {
		existing[entry.RawMessage] = true
	}

	// Add new entries
	for _, entry := range other.Globals {
		if !existing[entry.RawMessage] {
			db.Globals = append(db.Globals, entry)
			existing[entry.RawMessage] = true
			added++
		}
	}

	return added
}

// GetPlayerGlobals returns all globals for the specified player
func (db *EntropyDB) GetPlayerGlobals() []GlobalEntry {
	if db.PlayerName == "" {
		return db.Globals
	}

	var results []GlobalEntry
	for _, entry := range db.Globals {
		if strings.EqualFold(entry.PlayerName, db.PlayerName) {
			results = append(results, entry)
		}
	}
	return results
}

// GetPlayerHofs returns all Hall of Fame entries for the specified player
func (db *EntropyDB) GetPlayerHofs() []GlobalEntry {
	if db.PlayerName == "" {
		return db.GetHofEntries()
	}

	var results []GlobalEntry
	for _, entry := range db.Globals {
		if entry.IsHof && strings.EqualFold(entry.PlayerName, db.PlayerName) {
			results = append(results, entry)
		}
	}
	return results
}
