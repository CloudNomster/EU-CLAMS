// Package gui provides graphical user interface for EU-CLAMS
package gui

import (
	"eu-clams/internal/config"
	"eu-clams/internal/logger"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ConfigGUI manages the configuration GUI window
type ConfigGUI struct {
	app        fyne.App
	mainWindow fyne.Window
	config     config.Config
	log        *logger.Logger
	// Form fields
	playerNameEntry        *widget.Entry
	teamNameEntry          *widget.Entry
	dbPathEntry            *widget.Entry
	chatLogPathEntry       *widget.Entry
	enableScreenshotsCheck *widget.Check
	screenshotDirEntry     *widget.Entry
	screenshotDelayEntry   *widget.Entry
	gameWindowTitleEntry   *widget.Entry
	enableWebServerCheck   *widget.Check
	webServerPortEntry     *widget.Entry

	// Save callback
	onSaveCallback func(config config.Config)
}

// NewConfigGUI creates a new configuration GUI
func NewConfigGUI(log *logger.Logger, cfg config.Config) *ConfigGUI {
	window := fyne.CurrentApp().NewWindow("EU-CLAMS Configuration")

	gui := &ConfigGUI{
		app:        nil, // We don't need to store the app since we're using the current app
		mainWindow: window,
		config:     cfg,
		log:        log,
	}

	return gui
}

// SetSaveCallback sets the callback function to be called when configuration is saved
func (g *ConfigGUI) SetSaveCallback(callback func(config config.Config)) {
	g.onSaveCallback = callback
}

// Show displays the GUI window
func (g *ConfigGUI) Show() {
	g.createUI()
	g.mainWindow.Show()
}

// createUI creates the user interface components
func (g *ConfigGUI) createUI() {
	// Create form fields
	g.playerNameEntry = widget.NewEntry()
	g.playerNameEntry.SetText(g.config.PlayerName)
	g.playerNameEntry.SetPlaceHolder("Enter your character name")

	g.teamNameEntry = widget.NewEntry()
	g.teamNameEntry.SetText(g.config.TeamName)
	g.teamNameEntry.SetPlaceHolder("Enter your team name (optional)")

	g.dbPathEntry = widget.NewEntry()
	g.dbPathEntry.SetText(g.config.DatabasePath)
	g.dbPathEntry.SetPlaceHolder("Path to database file")

	g.chatLogPathEntry = widget.NewEntry()
	g.chatLogPathEntry.SetPlaceHolder("Path to EU chat log file")
	defaultChatLogPath := filepath.Join(os.Getenv("USERPROFILE"), "Documents", "Entropia Universe", "chat.log")
	if _, err := os.Stat(defaultChatLogPath); err == nil {
		g.chatLogPathEntry.SetText(defaultChatLogPath)
	}

	// Create screenshot related fields
	g.enableScreenshotsCheck = widget.NewCheck("", nil)
	g.enableScreenshotsCheck.SetChecked(g.config.EnableScreenshots)
	g.screenshotDirEntry = widget.NewEntry()
	g.screenshotDirEntry.SetText(g.config.ScreenshotDirectory)
	g.screenshotDirEntry.SetPlaceHolder("Path to screenshot directory")
	g.screenshotDelayEntry = widget.NewEntry()
	g.screenshotDelayEntry.SetText(fmt.Sprintf("%.1f", g.config.ScreenshotDelay))
	g.screenshotDelayEntry.SetPlaceHolder("0.6")
	g.gameWindowTitleEntry = widget.NewEntry()
	g.gameWindowTitleEntry.SetText(g.config.GameWindowTitle)
	g.gameWindowTitleEntry.SetPlaceHolder("Entropia Universe Client")
	// Create web server related fields
	g.enableWebServerCheck = widget.NewCheck("", nil)
	g.enableWebServerCheck.SetChecked(g.config.EnableWebServer)
	g.webServerPortEntry = widget.NewEntry()
	g.webServerPortEntry.SetText(strconv.Itoa(g.config.WebServerPort))
	g.webServerPortEntry.SetPlaceHolder("8080")

	// Create buttons for file selection
	dbPathButton := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err != nil || uri == nil {
				return
			}
			g.dbPathEntry.SetText(uri.URI().Path())
		}, g.mainWindow)
	})

	chatLogPathButton := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err != nil || uri == nil {
				return
			}
			g.chatLogPathEntry.SetText(uri.URI().Path())
		}, g.mainWindow)
	})

	screenshotDirButton := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			g.screenshotDirEntry.SetText(uri.Path())
		}, g.mainWindow)
	})

	// Create database path container with browse button
	dbPathContainer := container.NewBorder(nil, nil, nil, dbPathButton, g.dbPathEntry)
	chatLogPathContainer := container.NewBorder(nil, nil, nil, chatLogPathButton, g.chatLogPathEntry)
	screenshotDirContainer := container.NewBorder(nil, nil, nil, screenshotDirButton, g.screenshotDirEntry) // Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Player Name", Widget: g.playerNameEntry, HintText: "Your character name in Entropia Universe"},
			{Text: "Team Name", Widget: g.teamNameEntry, HintText: "Your team name (optional)"},
			{Text: "Database Path", Widget: dbPathContainer, HintText: "Where to store your globals database"},
			{Text: "Chat Log Path", Widget: chatLogPathContainer, HintText: "Path to Entropia Universe chat.log"},
			{Text: "Enable Screenshots", Widget: g.enableScreenshotsCheck, HintText: "Take screenshots for globals and HoFs"},
			{Text: "Screenshot Directory", Widget: screenshotDirContainer, HintText: "Where to save screenshots"},
			{Text: "Screenshot Delay", Widget: g.screenshotDelayEntry, HintText: "Delay in seconds before taking a screenshot (default: 0.6)"},
			{Text: "Game Window Title", Widget: g.gameWindowTitleEntry, HintText: "Beginning of Entropia Universe window title"},
			{Text: "Enable Web Server", Widget: g.enableWebServerCheck, HintText: "Start a web server to view statistics"},
			{Text: "Web Server Port", Widget: g.webServerPortEntry, HintText: "Port for the web server (default: 8080)"},
		}, OnSubmit: g.saveConfig,
		OnCancel: func() {
			g.mainWindow.Close()
		},
		SubmitText: "Save Configuration",
		CancelText: "Close",
	}
	// Main content
	content := container.NewVBox(
		widget.NewLabelWithStyle("EU-CLAMS Configuration", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		form,
	)

	g.mainWindow.SetContent(content)
	g.mainWindow.Resize(fyne.NewSize(600, 500))
	g.mainWindow.CenterOnScreen()
}

// saveConfig saves the configuration
func (g *ConfigGUI) saveConfig() { // Update configuration values from form fields
	g.config.PlayerName = g.playerNameEntry.Text
	g.config.TeamName = g.teamNameEntry.Text
	g.config.DatabasePath = g.dbPathEntry.Text
	g.config.EnableScreenshots = g.enableScreenshotsCheck.Checked
	g.config.ScreenshotDirectory = g.screenshotDirEntry.Text
	g.config.GameWindowTitle = g.gameWindowTitleEntry.Text
	g.config.EnableWebServer = g.enableWebServerCheck.Checked

	// Convert screenshot delay from string to float64
	screenshotDelay := 0.6 // Default delay
	if delay, err := strconv.ParseFloat(g.screenshotDelayEntry.Text, 64); err == nil && delay >= 0 {
		screenshotDelay = delay
	} else if g.screenshotDelayEntry.Text != "" {
		dialog.ShowError(fmt.Errorf("invalid screenshot delay: must be a positive number"), g.mainWindow)
		return
	}
	g.config.ScreenshotDelay = screenshotDelay

	// Convert web server port from string to int
	webServerPort := 8080 // Default port
	if port, err := strconv.Atoi(g.webServerPortEntry.Text); err == nil && port > 0 {
		webServerPort = port
	} else if g.webServerPortEntry.Text != "" {
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

	// Call the callback if defined
	if g.onSaveCallback != nil {
		g.onSaveCallback(g.config)
	}

	// Show success message
	dialog.ShowInformation("Success", "Configuration saved successfully", g.mainWindow)

	// Close the window
	g.mainWindow.Close()
}
