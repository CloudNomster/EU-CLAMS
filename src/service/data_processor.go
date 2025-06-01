package service

import (
	"eu-clams/internal/config"
	"eu-clams/internal/logger"
	"eu-clams/internal/storage"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DataProcessorService processes chat logs and extracts globals data
type DataProcessorService struct {
	*BaseService
	log          *logger.Logger
	config       config.Config
	db           *storage.EntropyDB
	chatLogPath  string
	progressChan chan float64
	stopChan     chan struct{}
	watchDelay   time.Duration
	lastGlobal   time.Time
	isImportMode bool // Flag to indicate import-only mode, no screenshots
}

// NewDataProcessorService creates a new DataProcessorService instance
func NewDataProcessorService(log *logger.Logger, cfg config.Config, chatLogPath string) *DataProcessorService {
	return &DataProcessorService{
		BaseService:  NewBaseService("DataProcessor"),
		log:          log,
		config:       cfg,
		chatLogPath:  chatLogPath,
		progressChan: make(chan float64, 1),
		stopChan:     make(chan struct{}),
		watchDelay:   time.Second * 1, // Check for changes every second
		lastGlobal:   time.Time{},
		isImportMode: false, // Default to monitoring mode which takes screenshots
	}
}

// Initialize initializes the service
func (s *DataProcessorService) Initialize() error {
	s.log.Info("DataProcessor service initializing...")

	// Ensure database directory exists
	dbPath := s.config.DatabasePath
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(filepath.Dir(os.Args[0]), dbPath)
		s.config.DatabasePath = dbPath
	}
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Load existing database or create new one
	var err error
	s.db, err = storage.LoadDatabase(dbPath, s.log)
	if err != nil {
		s.log.Error("Failed to load database: %v", err)
		// Create a new database
		s.log.Info("Creating new database")
		s.db = storage.NewEntropyDB(s.config.PlayerName, s.config.TeamName)
	}

	// Update player/team names in database if needed
	if s.config.PlayerName != "" && s.db.PlayerName != s.config.PlayerName {
		s.log.Info("Updating player name in database from '%s' to '%s'", s.db.PlayerName, s.config.PlayerName)
		s.db.PlayerName = s.config.PlayerName
	}
	if s.config.TeamName != "" && s.db.TeamName != s.config.TeamName {
		s.log.Info("Updating team name in database from '%s' to '%s'", s.db.TeamName, s.config.TeamName)
		s.db.TeamName = s.config.TeamName
	}

	return nil
}

// Run executes the service logic
func (s *DataProcessorService) Run() error {
	s.log.Info("DataProcessor service starting...")
	defer s.log.LogTiming("DataProcessor.Run")()

	// Validate inputs
	if s.chatLogPath == "" {
		s.log.Error("Chat log path is required")
		return fmt.Errorf("chat log path is required")
	}

	// Make sure we're not in import mode when running regular monitoring
	s.isImportMode = false

	// Initial processing of the log file
	if err := s.processLogFile(); err != nil {
		return err
	}

	// Start watching for changes
	go s.watchLogFile()

	// Wait for stop signal
	<-s.stopChan
	return nil
}

// SetProgressChannel sets the progress channel for reporting import progress
func (s *DataProcessorService) SetProgressChannel(ch chan float64) {
	s.progressChan = ch
}

// GetProgressChannel returns the progress channel
func (s *DataProcessorService) GetProgressChannel() chan float64 {
	return s.progressChan
}

// StopMonitoring stops monitoring the chat log file
func (s *DataProcessorService) StopMonitoring() {
	s.log.Info("Stopping monitoring")
	close(s.stopChan)
}

// processLogFile processes the chat log file from the last known position
func (s *DataProcessorService) processLogFile() error {
	if _, err := os.Stat(s.chatLogPath); os.IsNotExist(err) {
		return fmt.Errorf("chat log file not found: %s", s.chatLogPath)
	}

	// Get current file size to track progress
	file, err := os.Open(s.chatLogPath)
	if err != nil {
		return fmt.Errorf("failed to open chat log: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Store the current globals count before processing
	oldGlobalsCount := len(s.db.Globals)

	// If we haven't processed this file before, process it from the beginning
	if s.db.LastProcessedSize == 0 {
		s.log.Info("Processing entire chat log file: %s", s.chatLogPath)
		count, err := s.db.ProcessChatLog(s.chatLogPath, s.progressChan, s.log)
		if err != nil {
			return fmt.Errorf("failed to process chat log: %w", err)
		}
		s.log.Info("Processed %d global entries", count)
	} else if fileInfo.Size() > s.db.LastProcessedSize {
		// Process only new content
		s.log.Debug("Processing new entries in chat log")
		count, err := s.db.ProcessChatLogFromOffset(s.chatLogPath, s.db.LastProcessedSize, nil, s.log)
		if err != nil {
			return fmt.Errorf("failed to process new entries: %w", err)
		}
		if count > 0 {
			s.log.Debug("Processed %d new global entries", count)

			// Get the new globals that were added
			newGlobals := s.db.Globals[oldGlobalsCount:]
			s.HandleNewGlobals(newGlobals)
		}
	}

	// Save the database after processing
	dbPath := s.config.DatabasePath
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(filepath.Dir(os.Args[0]), dbPath)
	}
	return s.db.SaveDatabase(dbPath, s.log)
}

// watchLogFile continuously watches the chat log file for changes
func (s *DataProcessorService) watchLogFile() {
	ticker := time.NewTicker(s.watchDelay)
	defer ticker.Stop()

	s.log.Info("Started watching chat log for changes: %s", s.chatLogPath)

	for {
		select {
		case <-s.stopChan:
			s.log.Info("Stopping chat log watcher")
			return
		case <-ticker.C:
			if err := s.processLogFile(); err != nil {
				s.log.Error("Error processing chat log: %v", err)
			}
		}
	}
}

// Stop stops the service
func (s *DataProcessorService) Stop() error {
	s.log.Info("DataProcessor service stopping...")
	close(s.stopChan)
	return nil
}

// GetDatabase returns the database instance
func (s *DataProcessorService) GetDatabase() *storage.EntropyDB {
	return s.db
}

// ProcessLogOnly processes the log file once without starting monitoring
func (s *DataProcessorService) ProcessLogOnly() error {
	s.log.Info("Processing log file once: %s", s.chatLogPath)
	// Set import mode flag to disable screenshots
	s.isImportMode = true

	// Process the log file
	if err := s.processLogFile(); err != nil {
		return err
	}
	return nil
}
