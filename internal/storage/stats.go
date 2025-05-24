package storage

import (
	"eu-clams/internal/model"
)

// GetStatsData generates stats data for the current database
func (db *EntropyDB) GetStatsData() model.Stats {
	// Convert storage.GlobalEntry to model.GlobalEntry
	modelEntries := make([]model.GlobalEntry, 0, len(db.Globals))
	for _, entry := range db.Globals {
		if db.PlayerName != "" && entry.PlayerName != db.PlayerName {
			continue // Skip entries not related to the player
		}
		modelEntry := model.GlobalEntry{
			Timestamp:  entry.Timestamp.Format("2006-01-02 15:04:05"),
			Type:       entry.Type,
			PlayerName: entry.PlayerName,
			TeamName:   entry.TeamName,
			Target:     entry.Target,
			Value:      entry.Value,
			Location:   entry.Location,
			IsHof:      entry.IsHof,
		}
		modelEntries = append(modelEntries, modelEntry)
	}

	// Generate stats using the model function
	return model.GenerateStatsFromGlobals(modelEntries)
}
