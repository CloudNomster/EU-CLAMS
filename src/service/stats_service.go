package service

import (
	"eu-clams/internal/logger"
	"eu-clams/internal/stats"
	"eu-clams/internal/storage"
	"fmt"
)

// StatsService handles statistics generation and reporting
type StatsService struct {
	*BaseService
	log        *logger.Logger
	db         *storage.EntropyDB
	playerName string
	teamName   string
}

// NewStatsService creates a new StatsService instance
func NewStatsService(log *logger.Logger, db *storage.EntropyDB, playerName string, teamName string) *StatsService {
	return &StatsService{
		BaseService: NewBaseService("StatsService"),
		log:         log,
		db:          db,
		playerName:  playerName,
		teamName:    teamName,
	}
}

// Initialize initializes the service
func (s *StatsService) Initialize() error {
	s.log.Info("StatsService initializing...")

	// Validate player name
	if s.playerName == "" {
		return fmt.Errorf("player name is required for statistics")
	}

	// Make sure we have a database to work with
	if s.db == nil {
		return fmt.Errorf("database is required for statistics")
	}

	return nil
}

// Run executes the service logic
func (s *StatsService) Run() error {
	s.log.Info("StatsService starting...")
	defer s.log.LogTiming("StatsService.Run")()

	// Generate statistics
	s.log.Info("Generating statistics for player: %s", s.playerName)
	statsData := s.db.GetStatsData()
	// Format and print report
	statsReport := stats.FormatStatsReport(statsData, s.playerName, s.teamName)
	fmt.Println("\n--- PLAYER STATISTICS ---")
	fmt.Println(statsReport)

	return nil
}

// GenerateStats returns statistics data
func (s *StatsService) GenerateStats() stats.Stats {
	s.log.Info("Generating statistics for player: %s", s.playerName)
	return s.db.GetStatsData()
}

// FormatStatsReport formats a statistics report as a string
func (s *StatsService) FormatStatsReport(statsData stats.Stats) string {
	return stats.FormatStatsReport(statsData, s.playerName, s.teamName)
}

// Stop stops the service
func (s *StatsService) Stop() error {
	s.log.Info("StatsService stopping...")
	return nil
}
