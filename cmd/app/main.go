package main

import (
	"eu-clams/internal/config"
	"eu-clams/internal/gui"
	"eu-clams/internal/logger"
	"eu-clams/src/service"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

const (
	version = "0.1.0"
)

var log = logger.New()

func main() {
	// Define command-line flags
	configPath := flag.String("config", "", "Path to configuration file")
	logPath := flag.String("log", "", "Path to Entropia Universe chat log file")
	playerName := flag.String("player", "", "Your character name in Entropia Universe")
	teamName := flag.String("team", "", "Your team name in Entropia Universe")
	showStats := flag.Bool("stats", false, "Show statistics for your globals and HoFs")
	showVersion := flag.Bool("version", false, "Display version information")
	importLog := flag.Bool("import", false, "Import the chat log file without monitoring")
	monitor := flag.Bool("monitor", false, "Monitor chat log for changes (default true)")
	useCLI := flag.Bool("cli", false, "Use command-line interface instead of GUI")
	_ = flag.Bool("verbose", false, "Enable verbose logging") // Unused for now

	// Parse command-line flags
	flag.Parse()
	// Display version if requested
	if *showVersion {
		fmt.Printf("EU-CLAMS v%s\n", version)
		os.Exit(0)
	}

	log.Info("EU-CLAMS starting...")

	// Load configuration
	var cfg config.Config

	if *configPath != "" {
		var err error
		log.Info("Loading configuration from: %s", *configPath)
		cfg, err = config.LoadConfigFromFile(*configPath)
		if err != nil {
			log.Error("Failed to load config: %v", err)
			os.Exit(1)
		}
	} else {
		// Check for config.yaml in current directory
		defaultConfigPath := "config.yaml"
		if _, err := os.Stat(defaultConfigPath); err == nil {
			log.Info("Loading configuration from: %s", defaultConfigPath)
			cfg, err = config.LoadConfigFromFile(defaultConfigPath)
			if err != nil {
				log.Error("Failed to load config: %v", err)
				os.Exit(1)
			}
		} else {
			// Use default configuration
			log.Info("No configuration file found, using default configuration")
			cfg = config.NewDefaultConfig()
		}

		// Ensure database path is absolute
		if !filepath.IsAbs(cfg.DatabasePath) {
			cfg.DatabasePath = filepath.Join(filepath.Dir(os.Args[0]), cfg.DatabasePath)
		}
	}

	log.Info("App: %s v%s", cfg.AppName, cfg.Version)

	// Override player/team names from command line if provided
	if *playerName != "" {
		cfg.PlayerName = *playerName
		log.Info("Using player name from command line: %s", cfg.PlayerName)
	}
	if *teamName != "" {
		cfg.TeamName = *teamName
		log.Info("Using team name from command line: %s", cfg.TeamName)
	}

	// Use command-line interface if explicitly requested or if certain flags are set
	if *useCLI || *showStats || *importLog || *monitor {
		log.Info("Starting in CLI mode")
		// Continue with CLI mode
		// Determine log file path
		chatLogPath := *logPath
		if chatLogPath == "" {
			// Try to use default path if not specified
			defaultPath := filepath.Join(os.Getenv("USERPROFILE"), "Documents", "Entropia Universe", "chat.log")
			if _, err := os.Stat(defaultPath); err == nil {
				chatLogPath = defaultPath
				log.Info("Using default chat log path: %s", chatLogPath)
			} else {
				log.Error("No chat log path specified. Use -log flag.")
				fmt.Println("Usage: eu-tool -log <path-to-chat.log> -player <your-character-name> [-team <your-team-name>]")
				os.Exit(1)
			}
		}

		// Initialize data processor service
		dataProcessor := service.NewDataProcessorService(log, cfg, chatLogPath)
		if err := dataProcessor.Initialize(); err != nil {
			log.Error("Failed to initialize data processor: %v", err)
			os.Exit(1)
		}

		// Run the data processor
		if *importLog {
			// One-time import mode
			if err := dataProcessor.Run(); err != nil {
				log.Error("Failed to process data: %v", err)
				os.Exit(1)
			}
			dataProcessor.Stop()
		} else if *monitor || (!*importLog && !*showStats) {
			// Start monitoring in background
			go func() {
				if err := dataProcessor.Run(); err != nil {
					log.Error("Failed to start monitoring: %v", err)
					os.Exit(1)
				}
			}()

			log.Info("Press Ctrl+C to stop monitoring...")

			// Handle Ctrl+C gracefully
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			log.Info("Stopping services...")
			dataProcessor.Stop()
		}

		// Show statistics if requested
		if *showStats {
			if cfg.PlayerName == "" {
				log.Error("Player name is required for statistics. Use -player flag.")
				os.Exit(1)
			}

			statsService := service.NewStatsService(log, dataProcessor.GetDatabase(), cfg.PlayerName, cfg.TeamName)
			if err := statsService.Initialize(); err != nil {
				log.Error("Failed to initialize stats service: %v", err)
				os.Exit(1)
			}

			if err := statsService.Run(); err != nil {
				log.Error("Failed to generate statistics: %v", err)
				os.Exit(1)
			}
		}

		return // Exit after completing CLI operations
	}

	// By default, launch the GUI mode
	log.Info("Starting in GUI mode")
	mainGUI := gui.NewMainGUI(log, cfg)
	mainGUI.Show()
}
