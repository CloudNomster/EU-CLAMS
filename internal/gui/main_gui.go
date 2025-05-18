package gui

import (
	"eu-clams/internal/config"
	"eu-clams/internal/logger"
	"eu-clams/internal/storage"
	"eu-clams/src/service"
	"fmt"
	"os"
	"path/filepath"

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
	g.mainWindow.ShowAndRun()
}

// createUI creates the user interface components
func (g *MainGUI) createUI() {
	// Create status label
	g.statusLabel = widget.NewLabelWithStyle("Ready", fyne.TextAlignCenter, fyne.TextStyle{})
	// Create main action buttons
	configButton := widget.NewButtonWithIcon("Configure", theme.SettingsIcon(), g.showConfigDialog)
	g.monitorButton = widget.NewButtonWithIcon("Start Monitoring", theme.MediaPlayIcon(), g.toggleMonitoring)
	importButton := widget.NewButtonWithIcon("Import Log", theme.DownloadIcon(), g.importChatLog)
	statsButton := widget.NewButtonWithIcon("View Statistics", theme.DocumentIcon(), g.showStats)

	// Create button container
	buttonsContainer := container.New(layout.NewGridLayout(2), configButton,
		g.monitorButton,
		importButton,
		statsButton,
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
	// Main content
	content := container.NewVBox(
		widget.NewLabelWithStyle("EU-CLAMS v1.0", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		boxesContainer,
		buttonsContainer,
	)

	g.mainWindow.SetContent(content)
	g.mainWindow.Resize(fyne.NewSize(600, 400))
	g.mainWindow.CenterOnScreen()
}

// valueOrEmpty returns the value or a default if empty
func valueOrEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// showConfigDialog shows the configuration dialog
func (g *MainGUI) showConfigDialog() {
	configGUI := NewConfigGUI(g.log, g.config)
	configGUI.SetSaveCallback(func(newConfig config.Config) {
		g.config = newConfig
		g.log.Info("Updated configuration from GUI")

		// Update the info label
		infoText := fmt.Sprintf("Player: %s\nTeam: %s\n",
			valueOrEmpty(g.config.PlayerName, "Not set"),
			valueOrEmpty(g.config.TeamName, "Not set"))
		g.infoLabel.SetText(infoText)

		// Update services if needed
		if g.dataService != nil {
			// Reload with new config
		}
	})
	configGUI.Show()
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
func (g *MainGUI) startMonitoring() { // Validate configuration
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
	} // Update UI on the main thread first
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
		// Show progress dialog using the recommended approach
		progressContent := container.NewVBox(
			widget.NewLabel("Processing chat log..."),
			widget.NewProgressBar(),
		)
		progress := dialog.NewCustomWithoutButtons("Importing", progressContent, g.mainWindow)
		progress.Show()

		// Get the progress bar widget to update later
		progressBar := progressContent.Objects[1].(*widget.ProgressBar)

		// Process chat log asynchronously
		go func() {
			// Set up progress monitoring
			progressChan := make(chan float64, 10)
			go func() {
				for p := range progressChan { // Thread-safe UI update
					fyne.Do(func() {
						progressBar.SetValue(p / 100.0)
					})
				}
			}() // Run service - one-time import, don't start monitoring
			dataService.SetProgressChannel(progressChan)
			if err := dataService.ProcessLogOnly(); err != nil {
				g.log.Error("Import error: %v", err)
				// Use a channel to communicate back to the main thread
				errorChan := make(chan error, 1)
				errorChan <- err
				go func() {
					err := <-errorChan
					fyne.Do(func() {
						progress.Hide()
						dialog.ShowError(err, g.mainWindow)
					})
				}()
				return
			} // Signal success back to main thread
			doneChan := make(chan struct{}, 1)
			doneChan <- struct{}{}
			go func() {
				<-doneChan
				fyne.Do(func() {
					progress.Hide()
					dialog.ShowInformation("Success", "Chat log imported successfully", g.mainWindow)
					g.statusLabel.SetText("Import completed")
				})
			}()
		}()
	}, g.mainWindow)
}

// showStats shows the statistics
func (g *MainGUI) showStats() {
	// Validate configuration
	if g.config.PlayerName == "" {
		dialog.ShowError(fmt.Errorf("player name is required"), g.mainWindow)
		return
	}

	// Get database from existing service or create new one
	var db *storage.EntropyDB
	if g.dataService != nil {
		db = g.dataService.GetDatabase()
	} else {
		// Initialize data processor service just to get database
		tempService := service.NewDataProcessorService(g.log, g.config, "")
		if err := tempService.Initialize(); err != nil {
			dialog.ShowError(fmt.Errorf("failed to initialize data processor: %w", err), g.mainWindow)
			return
		}
		db = tempService.GetDatabase()
	}

	// Initialize stats service
	statsService := service.NewStatsService(g.log, db, g.config.PlayerName, g.config.TeamName)
	if err := statsService.Initialize(); err != nil {
		dialog.ShowError(fmt.Errorf("failed to initialize stats service: %w", err), g.mainWindow)
		return
	}

	// Generate stats text
	statsData := statsService.GenerateStats()
	statsText := statsService.FormatStatsReport(statsData)
	// Show stats dialog
	statsWindow := g.app.NewWindow("Statistics")

	statsScroll := container.NewScroll(widget.NewLabel(statsText))
	statsScroll.SetMinSize(fyne.NewSize(500, 400))

	closeButton := widget.NewButton("Close", func() {
		statsWindow.Close()
	})

	content := container.NewBorder(nil, container.NewHBox(layout.NewSpacer(), closeButton), nil, nil, statsScroll)

	statsWindow.SetContent(content)
	statsWindow.Resize(fyne.NewSize(600, 500))
	statsWindow.CenterOnScreen()
	statsWindow.Show()
}
