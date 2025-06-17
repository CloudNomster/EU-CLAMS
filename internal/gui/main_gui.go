package gui

import (
	"eu-clams/internal/config"
	"eu-clams/internal/logger"
	"eu-clams/internal/storage"
	"eu-clams/src/service"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MainGUI is the main application GUI
type MainGUI struct {
	app         fyne.App
	mainWindow  fyne.Window
	config      config.Config
	log         *logger.Logger
	dataService *service.DataProcessorService
	webService  *service.WebService // Track the web service instance

	// Status variables
	isMonitoring  bool
	statusLabel   *widget.Label
	monitorButton *widget.Button
	infoLabel     *widget.Label
}

// NewMainGUI creates a new main GUI
func NewMainGUI(log *logger.Logger, cfg config.Config) *MainGUI {
	a := GetApp()
	window := a.NewWindow("EU-CLAMS")

	gui := &MainGUI{
		app:          a,
		mainWindow:   window,
		config:       cfg,
		log:          log,
		isMonitoring: false,
	}

	return gui
}

// Show displays the main GUI window
func (g *MainGUI) Show() {
	g.createUI()

	// Set up cleanup handler for when the window is closed
	g.mainWindow.SetCloseIntercept(func() {
		g.Close()
		g.mainWindow.Close()
	})

	// Start the web server if enabled in the config
	if g.config.EnableWebServer {
		url, err := g.initWebServer()
		if err != nil {
			g.log.Error("Failed to start web server: %v", err)
		} else {
			g.log.Info("Web server started successfully at %s", url)
		}
	}

	g.mainWindow.ShowAndRun()
}

// createUI creates the user interface components
func (g *MainGUI) createUI() {
	// Create status label
	g.statusLabel = widget.NewLabelWithStyle("Ready", fyne.TextAlignCenter, fyne.TextStyle{})
	// Create main action buttons
	g.monitorButton = widget.NewButtonWithIcon("Start Monitoring", theme.MediaPlayIcon(), g.toggleMonitoring)
	importButton := widget.NewButtonWithIcon("Import Log", theme.DownloadIcon(), g.importChatLog)
	webServerButton := widget.NewButtonWithIcon("Open Webstats", theme.ComputerIcon(), func() { g.startWebServer(true) })

	// Create button container
	buttonsContainer := container.New(layout.NewGridLayout(2),
		g.monitorButton,
		importButton,
		webServerButton,
	)

	// Create info label
	infoText := fmt.Sprintf("Player: %s\nTeam: %s\n",
		valueOrEmpty(g.config.PlayerName, "Not set"),
		valueOrEmpty(g.config.TeamName, "Not set"))
	g.infoLabel = widget.NewLabel(infoText)

	// Create info box
	infoBox := container.NewVBox(
		widget.NewLabelWithStyle("Configuration", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		g.infoLabel,
	)

	// Create status box
	statusBox := container.NewVBox(
		widget.NewLabelWithStyle("Status", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		g.statusLabel,
	)

	// Create panels for the boxes
	infoPanel := widget.NewCard("", "", infoBox)
	statusPanel := widget.NewCard("", "", statusBox)

	// Create boxes container
	boxesContainer := container.New(layout.NewGridLayout(2),
		infoPanel,
		statusPanel,
	)

	// Create main content (Dashboard tab)
	dashboardContent := container.NewVBox(
		widget.NewLabelWithStyle("EU-CLAMS Dashboard", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		boxesContainer,
		buttonsContainer,
	)

	// Create configuration tab content
	configContent := g.createConfigTab()

	// Create statistics tab content
	statsContent := g.createStatsTab()

	// Create tabs container
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Dashboard", theme.HomeIcon(), dashboardContent),
		container.NewTabItemWithIcon("Configuration", theme.SettingsIcon(), configContent),
		container.NewTabItemWithIcon("Statistics", theme.DocumentIcon(), statsContent),
	)

	g.mainWindow.SetContent(tabs)
	g.mainWindow.Resize(fyne.NewSize(600, 500))
	g.mainWindow.CenterOnScreen()
}

// valueOrEmpty returns the value or a default if empty
func valueOrEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// toggleMonitoring starts or stops monitoring
func (g *MainGUI) toggleMonitoring() {
	if g.isMonitoring {
		g.stopMonitoring()
		g.monitorButton.SetText("Start Monitoring")
		g.monitorButton.SetIcon(theme.MediaPlayIcon())
	} else {
		g.startMonitoring()
		g.monitorButton.SetText("Stop Monitoring")
		g.monitorButton.SetIcon(theme.MediaStopIcon())
	}
}

// startMonitoring starts monitoring the chat log
func (g *MainGUI) startMonitoring() {
	// Validate configuration
	if g.config.PlayerName == "" {
		dialog.ShowError(fmt.Errorf("player name is required"), g.mainWindow)
		return
	}

	// Determine chat log path
	chatLogPath := ""
	defaultPath := filepath.Join(os.Getenv("USERPROFILE"), "Documents", "Entropia Universe", "chat.log")
	if _, err := os.Stat(defaultPath); err == nil {
		chatLogPath = defaultPath
		g.log.Info("Using default chat log path: %s", chatLogPath)
	} else {
		// Show dialog to select chat log
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err != nil || uri == nil {
				dialog.ShowError(fmt.Errorf("chat log file is required"), g.mainWindow)
				return
			}
			chatLogPath = uri.URI().Path()
			g.startMonitoringWithPath(chatLogPath)
		}, g.mainWindow)
		return
	}

	g.startMonitoringWithPath(chatLogPath)
}

// startMonitoringWithPath starts monitoring with the specified chat log path
func (g *MainGUI) startMonitoringWithPath(chatLogPath string) {
	// Initialize data processor service
	g.dataService = service.NewDataProcessorService(g.log, g.config, chatLogPath)
	if err := g.dataService.Initialize(); err != nil {
		dialog.ShowError(fmt.Errorf("failed to initialize data processor: %w", err), g.mainWindow)
		return
	}
	// Update UI on the main thread first
	g.statusLabel.SetText("Monitoring chat log...")

	// Start monitoring asynchronously
	go func() {
		if err := g.dataService.Run(); err != nil {
			g.log.Error("Monitoring error: %v", err)
			// Keep it simple - let the main thread handle UI updates after goroutine completes
			g.isMonitoring = false
		}
	}()

	g.isMonitoring = true
}

// stopMonitoring stops monitoring the chat log
func (g *MainGUI) stopMonitoring() {
	if g.dataService != nil {
		g.dataService.StopMonitoring()
		g.dataService = nil
	}

	g.statusLabel.SetText("Monitoring stopped")
	g.isMonitoring = false
}

// importChatLog shows dialog to import chat log
func (g *MainGUI) importChatLog() {
	// Show file open dialog
	dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
		if err != nil || uri == nil {
			return
		}

		chatLogPath := uri.URI().Path()

		// Validate configuration
		if g.config.PlayerName == "" {
			dialog.ShowError(fmt.Errorf("player name is required"), g.mainWindow)
			return
		}

		// Initialize data processor service for import only
		dataService := service.NewDataProcessorService(g.log, g.config, chatLogPath)
		if err := dataService.Initialize(); err != nil {
			dialog.ShowError(fmt.Errorf("failed to initialize data processor: %w", err), g.mainWindow)
			return
		}
		// Show simple message that import is in progress
		importDialog := dialog.NewCustomWithoutButtons(
			"Importing",
			container.NewVBox(widget.NewLabel("Processing chat log... Please wait.")),
			g.mainWindow,
		)
		importDialog.Show() // Process chat log asynchronously
		go func() {
			// Run service - one-time import, don't start monitoring
			// Set progress channel to nil to disable progress updates
			dataService.SetProgressChannel(nil)
			if err := dataService.ProcessLogOnly(); err != nil {
				g.log.Error("Import error: %v", err)
				// Use a channel to communicate back to the main thread
				errorChan := make(chan error, 1)
				errorChan <- err
				go func() {
					err := <-errorChan
					fyne.Do(func() {
						importDialog.Hide()
						dialog.ShowError(err, g.mainWindow)
					})
				}()
				return
			}

			// Signal success back to main thread
			doneChan := make(chan struct{}, 1)
			doneChan <- struct{}{}
			go func() {
				<-doneChan
				fyne.Do(func() {
					importDialog.Hide()
					dialog.ShowInformation("Success", "Chat log imported successfully", g.mainWindow)
					g.statusLabel.SetText("Import completed")
				})
			}()
		}()
	}, g.mainWindow)
}

// initWebServer initializes and starts the web server if it's not already running
func (g *MainGUI) initWebServer() (string, error) {
	// If web service is already running, just return its URL
	if g.webService != nil {
		return fmt.Sprintf("http://localhost:%d", g.config.WebServerPort), nil
	}

	// Check if a web service for this port is already registered (started by CLI mode)
	webPort := g.config.WebServerPort
	if webPort <= 0 {
		webPort = 8080 // Default port
	}

	if existingService := service.GetWebServiceByPort(webPort); existingService != nil {
		g.log.Info("Found existing web service on port %d, reusing it", webPort)
		g.webService = existingService
		return fmt.Sprintf("http://localhost:%d", webPort), nil
	}

	// Validate configuration
	if g.config.PlayerName == "" {
		return "", fmt.Errorf("player name is required")
	}

	// Get database from existing service or create new one
	var db *storage.EntropyDB
	if g.dataService != nil {
		db = g.dataService.GetDatabase()
	} else {
		// Initialize data processor service just to get database
		tempService := service.NewDataProcessorService(g.log, g.config, "")
		if err := tempService.Initialize(); err != nil {
			return "", fmt.Errorf("failed to initialize data processor: %w", err)
		}
		db = tempService.GetDatabase()
	}
	// Port is already set above
	// No need to redefine webPort here

	// Initialize web service if it's enabled in the config
	g.webService = service.NewWebService(g.log, db, g.config.PlayerName, g.config.TeamName, webPort)
	if err := g.webService.Initialize(); err != nil {
		g.webService = nil
		return "", fmt.Errorf("failed to initialize web server: %w", err)
	}

	// Start the web server in the background
	go func() {
		g.log.Info("Starting web server on port %d", webPort)
		if err := g.webService.Run(); err != nil {
			g.log.Error("Web server error: %v", err)
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("web server error: %w", err), g.mainWindow)
			})
			g.webService = nil // Clear the reference if it fails to start
		}
	}()

	// Return the URL to the web stats
	return fmt.Sprintf("http://localhost:%d", webPort), nil
}

// startWebServer launches the web browser to the web stats
func (g *MainGUI) startWebServer(showDialog bool) {
	// Check if web server is enabled in the config
	if !g.config.EnableWebServer {
		// If not enabled, show an error dialog
		dialog.ShowError(fmt.Errorf("web server is not enabled in the configuration"), g.mainWindow)
		return
	}

	// Start the web server if it's not already running
	var url string
	if g.webService == nil {
		var err error
		url, err = g.initWebServer()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to start web server: %w", err), g.mainWindow)
			return
		}
	} else {
		// Use port from config for existing web service
		webPort := g.config.WebServerPort
		if webPort <= 0 {
			webPort = 8080 // Default port
		}
		url = fmt.Sprintf("http://localhost:%d", webPort)
	}

	// If showing dialog was requested, show information dialog
	if showDialog {
		info := fmt.Sprintf("Opening web statistics at %s", url)
		dialog.ShowInformation("Web Statistics", info, g.mainWindow)
	}

	// Try to open the browser
	openBrowser(url)
}

// openBrowser opens the default browser to the specified URL
func openBrowser(url string) {
	var err error

	switch {
	case hasCommand("rundll32"):
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case hasCommand("explorer"):
		err = exec.Command("explorer", url).Start()
	default:
		err = fmt.Errorf("could not find browser command")
	}

	if err != nil {
		fmt.Printf("Error opening browser: %v\n", err)
	}
}

// hasCommand checks if a command exists
func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// createConfigTab creates the content for the configuration tab
func (g *MainGUI) createConfigTab() fyne.CanvasObject {
	// Create form fields
	playerNameEntry := widget.NewEntry()
	playerNameEntry.SetText(g.config.PlayerName)
	playerNameEntry.SetPlaceHolder("Enter your character name")

	teamNameEntry := widget.NewEntry()
	teamNameEntry.SetText(g.config.TeamName)
	teamNameEntry.SetPlaceHolder("Enter your team name (optional)")

	dbPathEntry := widget.NewEntry()
	dbPathEntry.SetText(g.config.DatabasePath)
	dbPathEntry.SetPlaceHolder("Path to database file")

	chatLogPathEntry := widget.NewEntry()
	chatLogPathEntry.SetPlaceHolder("Path to EU chat log file")
	defaultChatLogPath := filepath.Join(os.Getenv("USERPROFILE"), "Documents", "Entropia Universe", "chat.log")
	if _, err := os.Stat(defaultChatLogPath); err == nil {
		chatLogPathEntry.SetText(defaultChatLogPath)
	}

	// Create screenshot related fields
	enableScreenshotsCheck := widget.NewCheck("", nil)
	enableScreenshotsCheck.SetChecked(g.config.EnableScreenshots)
	screenshotDirEntry := widget.NewEntry()
	screenshotDirEntry.SetText(g.config.ScreenshotDirectory)
	screenshotDirEntry.SetPlaceHolder("Path to screenshot directory")

	screenshotDelayEntry := widget.NewEntry()
	screenshotDelayEntry.SetText(fmt.Sprintf("%.1f", g.config.ScreenshotDelay))
	screenshotDelayEntry.SetPlaceHolder("0.6")

	gameWindowTitleEntry := widget.NewEntry()
	gameWindowTitleEntry.SetText(g.config.GameWindowTitle)
	gameWindowTitleEntry.SetPlaceHolder("Entropia Universe Client")

	// Create web server related fields
	enableWebServerCheck := widget.NewCheck("", nil)
	enableWebServerCheck.SetChecked(g.config.EnableWebServer)

	webServerPortEntry := widget.NewEntry()
	webServerPortEntry.SetText(strconv.Itoa(g.config.WebServerPort))
	webServerPortEntry.SetPlaceHolder("8080")

	// Create buttons for file selection
	dbPathButton := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err != nil || uri == nil {
				return
			}
			dbPathEntry.SetText(uri.URI().Path())
		}, g.mainWindow)
	})

	chatLogPathButton := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err != nil || uri == nil {
				return
			}
			chatLogPathEntry.SetText(uri.URI().Path())
		}, g.mainWindow)
	})

	screenshotDirButton := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			screenshotDirEntry.SetText(uri.Path())
		}, g.mainWindow)
	})
	// Create containers with browse buttons
	dbPathContainer := container.NewBorder(nil, nil, nil, dbPathButton, dbPathEntry)
	chatLogPathContainer := container.NewBorder(nil, nil, nil, chatLogPathButton, chatLogPathEntry)
	screenshotDirContainer := container.NewBorder(nil, nil, nil, screenshotDirButton, screenshotDirEntry)

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Player Name", Widget: playerNameEntry, HintText: "Your character name in Entropia Universe"},
			{Text: "Team Name", Widget: teamNameEntry, HintText: "Your team name (optional)"},
			{Text: "Database Path", Widget: dbPathContainer, HintText: "Where to store your globals database"},
			{Text: "Chat Log Path", Widget: chatLogPathContainer, HintText: "Path to Entropia Universe chat.log"},
			{Text: "Enable Screenshots", Widget: enableScreenshotsCheck, HintText: "Take screenshots for globals and HoFs"},
			{Text: "Screenshot Directory", Widget: screenshotDirContainer, HintText: "Where to save screenshots"},
			{Text: "Screenshot Delay", Widget: screenshotDelayEntry, HintText: "Delay in seconds before taking a screenshot (default: 0.6)"},
			{Text: "Game Window Title", Widget: gameWindowTitleEntry, HintText: "Beginning of Entropia Universe window title"},
			{Text: "Enable Web Server", Widget: enableWebServerCheck, HintText: "Start a web server to view statistics"}, {Text: "Web Server Port", Widget: webServerPortEntry, HintText: "Port for the web server (default: 8080)"},
		},
		OnSubmit: func() {
			// Update configuration values from form fields
			g.config.PlayerName = playerNameEntry.Text
			g.config.TeamName = teamNameEntry.Text
			g.config.DatabasePath = dbPathEntry.Text
			g.config.EnableScreenshots = enableScreenshotsCheck.Checked
			g.config.ScreenshotDirectory = screenshotDirEntry.Text
			g.config.GameWindowTitle = gameWindowTitleEntry.Text
			g.config.EnableWebServer = enableWebServerCheck.Checked

			// Convert screenshot delay from string to float64
			screenshotDelay := 0.6 // Default delay
			if delay, err := strconv.ParseFloat(screenshotDelayEntry.Text, 64); err == nil && delay >= 0 {
				screenshotDelay = delay
			} else if screenshotDelayEntry.Text != "" {
				dialog.ShowError(fmt.Errorf("invalid screenshot delay: must be a positive number"), g.mainWindow)
				return
			}
			g.config.ScreenshotDelay = screenshotDelay

			// Convert web server port from string to int
			webServerPort := 8080 // Default port
			if port, err := strconv.Atoi(webServerPortEntry.Text); err == nil && port > 0 {
				webServerPort = port
			} else if webServerPortEntry.Text != "" {
				dialog.ShowError(fmt.Errorf("invalid web server port: must be a positive number"), g.mainWindow)
				return
			}
			g.config.WebServerPort = webServerPort

			// Save to file
			err := g.config.SaveConfigToFile("config.yaml")
			if err != nil {
				dialog.ShowError(err, g.mainWindow)
				g.log.Error("Failed to save configuration: %v", err)
				return
			}

			g.log.Info("Configuration saved successfully")

			// Update the info label in the dashboard
			infoText := fmt.Sprintf("Player: %s\nTeam: %s\n",
				valueOrEmpty(g.config.PlayerName, "Not set"),
				valueOrEmpty(g.config.TeamName, "Not set"))
			g.infoLabel.SetText(infoText)

			// Update services if needed
			if g.dataService != nil {
				// Reload with new config
			}

			// Show success message
			dialog.ShowInformation("Success", "Configuration saved successfully", g.mainWindow)
		},
		SubmitText: "Save Configuration",
	}

	// Main content
	return container.NewVBox(
		widget.NewLabelWithStyle("EU-CLAMS Configuration", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		form,
	)
}

// createStatsTab creates the content for the statistics tab
func (g *MainGUI) createStatsTab() fyne.CanvasObject {
	// Create a label for statistics content
	statsLabel := widget.NewLabel("Click 'Refresh Statistics' to view the latest statistics")

	// Create a scroll container for the stats text
	statsScroll := container.NewScroll(statsLabel)
	statsScroll.SetMinSize(fyne.NewSize(500, 400))

	// Create a refresh button
	refreshButton := widget.NewButtonWithIcon("Refresh Statistics", theme.ViewRefreshIcon(), func() {
		// Create a progress dialog using the recommended approach
		progressContent := container.NewVBox(
			widget.NewLabel("Loading statistics data..."),
			widget.NewProgressBar(),
		)
		progressDialog := dialog.NewCustomWithoutButtons("Statistics", progressContent, g.mainWindow)
		progressDialog.Show()

		// Process in background
		go func() {
			// Validate configuration first
			if g.config.PlayerName == "" {
				fyne.Do(func() {
					progressDialog.Hide()
					statsLabel.SetText("Error: Player name is required in configuration")
				})
				return
			}

			// Get database
			var db *storage.EntropyDB
			if g.dataService != nil {
				db = g.dataService.GetDatabase()
			} else {
				// Initialize data processor service just to get database
				tempService := service.NewDataProcessorService(g.log, g.config, "")
				if err := tempService.Initialize(); err != nil {
					g.log.Error("Failed to initialize data processor: %v", err)
					fyne.Do(func() {
						progressDialog.Hide()
						statsLabel.SetText("Error: Failed to initialize data processor")
					})
					return
				}
				db = tempService.GetDatabase()
			}

			// Initialize stats service
			statsService := service.NewStatsService(g.log, db, g.config.PlayerName, g.config.TeamName)
			if err := statsService.Initialize(); err != nil {
				g.log.Error("Failed to initialize stats service: %v", err)
				fyne.Do(func() {
					progressDialog.Hide()
					statsLabel.SetText("Error: Failed to initialize stats service")
				})
				return
			}

			// Generate stats text
			statsData := statsService.GenerateStats()
			statsText := statsService.FormatStatsReport(statsData)

			// Update the stats label on the main thread
			fyne.Do(func() {
				progressDialog.Hide()
				statsLabel.SetText(statsText)
			})
		}()
	})

	// Create the main content with a refresh button at the bottom
	content := container.NewVBox(
		widget.NewLabelWithStyle("EU-CLAMS Statistics", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		statsScroll,
		container.NewHBox(layout.NewSpacer(), refreshButton),
	)

	return content
}

// Close properly cleans up resources before closing the application
func (g *MainGUI) Close() {
	// Stop the monitoring if active
	if g.isMonitoring && g.dataService != nil {
		g.stopMonitoring()
	}

	// Stop the web service if it's running
	if g.webService != nil {
		g.log.Info("Stopping web service...")
		if err := g.webService.Stop(); err != nil {
			g.log.Error("Error stopping web service: %v", err)
		}
		g.webService = nil
	}

	g.log.Info("Application closing")
}
