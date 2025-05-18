// Package gui provides graphical user interface for EU-CLAMS
package gui

import (
	"eu-clams/internal/config"
	"eu-clams/internal/logger"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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
	playerNameEntry  *widget.Entry
	teamNameEntry    *widget.Entry
	dbPathEntry      *widget.Entry
	chatLogPathEntry *widget.Entry

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

	// Create database path container with browse button
	dbPathContainer := container.NewBorder(nil, nil, nil, dbPathButton, g.dbPathEntry)
	chatLogPathContainer := container.NewBorder(nil, nil, nil, chatLogPathButton, g.chatLogPathEntry)

	// Create buttons
	saveButton := widget.NewButtonWithIcon("Save Configuration", theme.DocumentSaveIcon(), g.saveConfig)
	cancelButton := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		g.mainWindow.Close()
	})

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Player Name", Widget: g.playerNameEntry, HintText: "Your character name in Entropia Universe"},
			{Text: "Team Name", Widget: g.teamNameEntry, HintText: "Your team name (optional)"},
			{Text: "Database Path", Widget: dbPathContainer, HintText: "Where to store your globals database"},
			{Text: "Chat Log Path", Widget: chatLogPathContainer, HintText: "Path to Entropia Universe chat.log"},
		},
		OnSubmit: g.saveConfig,
		OnCancel: func() {
			g.mainWindow.Close()
		},
	}

	// Create buttons container
	buttons := container.New(layout.NewHBoxLayout(),
		layout.NewSpacer(),
		cancelButton,
		saveButton,
	)

	// Main content
	content := container.NewVBox(
		widget.NewLabelWithStyle("EU-CLAMS Configuration", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		form,
		buttons,
	)

	g.mainWindow.SetContent(content)
	g.mainWindow.Resize(fyne.NewSize(600, 400))
	g.mainWindow.CenterOnScreen()
}

// saveConfig saves the configuration
func (g *ConfigGUI) saveConfig() { // Update configuration values from form fields
	g.config.PlayerName = g.playerNameEntry.Text
	g.config.TeamName = g.teamNameEntry.Text
	g.config.DatabasePath = g.dbPathEntry.Text

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

	dialog.ShowInformation("Success", "Configuration saved successfully", g.mainWindow)
}
