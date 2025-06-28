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
	version = "0.1.10"
)

var log = logger.New()

func main() { // Define command-line flags
	configPath := flag.String("config", "", "Path to configuration file")
	logPath := flag.String("log", "", "Path to Entropia Universe chat log file")
	playerName := flag.String("player", "", "Your character name in Entropia Universe")
	teamName := flag.String("team", "", "Your team name in Entropia Universe")
	showStats := flag.Bool("stats", false, "Show statistics for your globals and HoFs")
	showVersion := flag.Bool("version", false, "Display version information")
	importLog := flag.Bool("import", false, "Import the chat log file without monitoring")
	monitor := flag.Bool("monitor", false, "Monitor chat log for changes (default true)")
	useCLI := flag.Bool("cli", false, "Use command-line interface instead of GUI")
	enableScreenshots := flag.Bool("screenshots", true, "Enable screenshots for globals and HoFs")
	screenshotDir := flag.String("screenshot-dir", "./data/screenshots", "Directory to save screenshots")
	gameWindow := flag.String("game-window", "Entropia Universe Client", "Game window title")
	webServer := flag.Bool("web", false, "Start a web server to view statistics")
	webPort := flag.Int("web-port", 8080, "Port for the web server")
	verbose := flag.Bool("verbose", false, "Enable verbose (debug) logging")

	// Parse command-line flags
	flag.Parse()
	// Display version if requested
	if *showVersion {
		fmt.Printf("EU-CLAMS v%s\n", version)
		os.Exit(0)
	}

	// Initialize logger with appropriate debug level
	if *verbose {
		log = logger.NewWithDebug()
		log.Debug("Debug logging enabled")
	}
	log.Info("EU-CLAMS starting...")
	// Load configuration using the centralized manager
	configManager := config.NewManager()
	// Set up command-line overrides
	overrides := config.CommandLineOverrides{
		ConfigPath:          configPath,
		PlayerName:          playerName,
		TeamName:            teamName,
		ChatLogPath:         logPath,
		EnableScreenshots:   enableScreenshots,
		ScreenshotDirectory: screenshotDir,
		GameWindowTitle:     gameWindow,
		EnableWebServer:     webServer,
		WebServerPort:       webPort,
	}

	// Load config with overrides
	cfg, err := configManager.LoadConfig(*configPath, overrides)
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	actualConfigPath := configManager.GetConfigPath()
	log.Info("Configuration loaded from: %s", actualConfigPath)

	log.Info("App: %s v%s", cfg.AppName, version)

	// Log effective configuration
	if cfg.PlayerName != "" {
		log.Info("Player name: %s", cfg.PlayerName)
	}
	if cfg.TeamName != "" {
		log.Info("Team name: %s", cfg.TeamName)
	}
	log.Info("Screenshot capture: %v", cfg.EnableScreenshots)
	log.Info("Screenshot directory: %s", cfg.ScreenshotDirectory)
	log.Info("Game window title: %s", cfg.GameWindowTitle)
	// Use command-line interface if explicitly requested or if certain flags are set
	if *useCLI || *showStats || *importLog || *monitor {
		log.Info("Starting in CLI mode")
		// Continue with CLI mode
		// Determine log file path from config, command line, or auto-detection
		chatLogPath := cfg.ChatLogPath
		if chatLogPath == "" {
			// Try to use default path if not specified in config
			defaultPath := filepath.Join(os.Getenv("USERPROFILE"), "Documents", "Entropia Universe", "chat.log")
			if _, err := os.Stat(defaultPath); err == nil {
				chatLogPath = defaultPath
				log.Info("Using default chat log path: %s", chatLogPath)
			} else {
				log.Error("No chat log path specified in config or command line, and default path not found. Use -log flag or add chat_log_path to config.yaml")
				fmt.Println("Usage: eu-tool -log <path-to-chat.log> -player <your-character-name> [-team <your-team-name>]")
				fmt.Println("Or add 'chat_log_path: <path>' to your config.yaml file")
				os.Exit(1)
			}
		} else {
			log.Info("Using chat log path from config: %s", chatLogPath)
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

			// Don't wait for Ctrl+C here, we'll do that after starting all services
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
		} // Start web server if requested via flag or config
		var webService *service.WebService
		// Determine if web server should be started (from config or command line flag)
		startWebServer := *webServer || cfg.EnableWebServer
		webServerPort := *webPort
		if cfg.EnableWebServer && !*webServer {
			// Use config port if web server was not explicitly requested via command line
			webServerPort = cfg.WebServerPort
		}

		if startWebServer {
			if cfg.PlayerName == "" {
				log.Error("Player name is required for web server. Use -player flag.")
				os.Exit(1)
			}

			log.Info("Starting web server on port %d...", webServerPort)
			webService = service.NewWebService(log, dataProcessor.GetDatabase(), cfg.PlayerName, cfg.TeamName, webServerPort)
			if err := webService.Initialize(); err != nil {
				log.Error("Failed to initialize web service: %v", err)
				os.Exit(1)
			} // Always start the web server in background
			go func() {
				if err := webService.Run(); err != nil {
					log.Error("Web server error: %v", err)
					os.Exit(1)
				}
			}()
		} // If we have any background services running, wait for Ctrl+C
		if *monitor || startWebServer || (!*importLog && !*showStats) {
			log.Info("Press Ctrl+C to stop services...")

			// Handle Ctrl+C gracefully
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			log.Info("Stopping services...")
			dataProcessor.Stop()
			if webService != nil {
				webService.Stop()
			}
		}

		return // Exit after completing CLI operations
	}
	// By default, launch the GUI mode
	log.Info("Starting in GUI mode")
	// If web server was started via command line flag, don't let GUI start it again
	if *webServer {
		// Override the config setting to prevent GUI from starting another web server
		cfg.EnableWebServer = false
		log.Info("Web server already started via command line, disabling automatic start in GUI")
	}

	mainGUI := gui.NewMainGUI(log, cfg)
	mainGUI.SetConfigManager(configManager)
	mainGUI.Show()
}
