package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// NewApp creates a new Fyne application
func NewApp() fyne.App {
	return app.NewWithID("com.github.eu-tool")
}

// GetApp returns the current Fyne application if it exists or creates a new one
func GetApp() fyne.App {
	return app.NewWithID("com.github.eu-tool")
}
