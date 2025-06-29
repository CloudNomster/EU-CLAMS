package service

import (
	"encoding/json"
	"eu-clams/internal/logger"
	"eu-clams/internal/model"
	"eu-clams/internal/storage"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebService provides a web interface for viewing statistics and events
type WebService struct {
	*BaseService
	log         *logger.Logger
	db          *storage.EntropyDB
	playerName  string
	teamName    string
	port        int
	server      *http.Server
	clients     map[*websocket.Conn]bool
	clientsLock sync.Mutex
	// Add a map to store write mutexes for each connection
	writeLockers map[*websocket.Conn]*sync.Mutex
	upgrader     websocket.Upgrader
	templates    *template.Template
	templateDir  string
	staticDir    string
}

// NewWebService creates a new WebService instance
func NewWebService(log *logger.Logger, db *storage.EntropyDB, playerName, teamName string, port int) *WebService {
	return &WebService{
		BaseService:  NewBaseService("WebService"),
		log:          log,
		db:           db,
		playerName:   playerName,
		teamName:     teamName,
		port:         port,
		clients:      make(map[*websocket.Conn]bool),
		writeLockers: make(map[*websocket.Conn]*sync.Mutex),
		upgrader:     websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		templateDir:  getTemplateDir(),
		staticDir:    getStaticDir(),
	}
}

// getTemplateDir returns the path to the templates directory
func getTemplateDir() string {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return "templates"
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "templates")
}

// getStaticDir returns the path to the static files directory
func getStaticDir() string {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return "templates/static"
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "templates/static")
}

// Initialize initializes the web service
func (s *WebService) Initialize() error {
	s.log.Info("WebService initializing...")

	// Validate required fields
	if s.db == nil {
		return fmt.Errorf("database is required for web service")
	}

	// Check if template directory exists
	if _, err := os.Stat(s.templateDir); os.IsNotExist(err) {
		// Try working directory
		cwd, _ := os.Getwd()
		s.templateDir = filepath.Join(cwd, "templates")
		s.staticDir = filepath.Join(cwd, "templates/static")
	}

	s.log.Info("Using template directory: %s", s.templateDir)

	// Initialize templates
	var err error
	templatePath := filepath.Join(s.templateDir, "index.html")
	s.templates, err = template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Register with the service registry
	RegisterWebService(fmt.Sprintf("web_%d", s.port), s)

	return nil
}

// Run starts the web server
func (s *WebService) Run() error {
	s.log.Info("WebService starting on port %d...", s.port)
	defer s.log.LogTiming("WebService.Run")()

	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/globals", s.handleGlobals)
	mux.HandleFunc("/api/hofs", s.handleHofs)
	mux.HandleFunc("/ws", s.handleWebSocket)

	// Serve static files
	staticHandler := http.FileServer(http.Dir(s.staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))

	// Create server
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	// Start the server
	s.log.Info("Web UI available at http://localhost:%d", s.port)
	return s.server.ListenAndServe()
}

// Stop stops the web server
func (s *WebService) Stop() error {
	s.log.Info("WebService stopping...")
	// Close all websocket connections
	s.clientsLock.Lock()
	for client := range s.clients {
		client.Close()
		delete(s.writeLockers, client)
	}
	s.clientsLock.Unlock()

	// Unregister from the service registry
	UnregisterWebService(fmt.Sprintf("web_%d", s.port))

	// Shutdown the server
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// handleIndex handles the index page
func (s *WebService) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Generate stats - always get fresh data from the database
	statsData := s.db.GetStatsData()
	allGlobals := s.db.GetPlayerGlobals()
	allHofs := s.db.GetPlayerHofs()

	// Limit to 10 entries
	var globals, hofs []storage.GlobalEntry
	if len(allGlobals) > 10 {
		globals = allGlobals[:10]
	} else {
		globals = allGlobals
	}

	if len(allHofs) > 10 {
		hofs = allHofs[:10]
	} else {
		hofs = allHofs
	}
	// Prepare template data
	data := map[string]interface{}{
		"PlayerName": s.playerName,
		"TeamName":   s.teamName,
		"Stats":      statsData,
		"Globals":    globals,
		"Hofs":       hofs,
		"Generated":  time.Now().UTC().Format(time.RFC3339),
	}
	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "text/html")

	// Render the template
	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		s.log.Error("Failed to render template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleStats handles the stats API endpoint
func (s *WebService) handleStats(w http.ResponseWriter, r *http.Request) {
	// Always get fresh stats from the database
	statsData := s.db.GetStatsData()

	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(statsData)
}

// handleGlobals handles the globals API endpoint
func (s *WebService) handleGlobals(w http.ResponseWriter, r *http.Request) {
	limit := 10 // Default to 10 to match the initial page load
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Always get fresh globals from the database
	globals := s.db.GetPlayerGlobals()
	if len(globals) > limit {
		globals = globals[:limit] // Take first 10 (already sorted newest first)
	}

	// Convert to JSON-friendly objects with ISO8601 UTC timestamps
	jsonGlobals := make([]model.GlobalEntryJSON, len(globals))
	for i, g := range globals {
		jsonGlobals[i] = model.GlobalEntryJSON{
			Timestamp:  g.Timestamp.UTC().Format(time.RFC3339),
			Type:       g.Type,
			PlayerName: g.PlayerName,
			TeamName:   g.TeamName,
			Target:     g.Target,
			Value:      g.Value,
			Location:   g.Location,
			IsHof:      g.IsHof,
			RawMessage: g.RawMessage,
		}
	}

	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(jsonGlobals)
}

// handleHofs handles the HOFs API endpoint
func (s *WebService) handleHofs(w http.ResponseWriter, r *http.Request) {
	limit := 10 // Default to 10 to match the initial page load
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Always get fresh HOFs from the database
	hofs := s.db.GetPlayerHofs()
	if len(hofs) > limit {
		hofs = hofs[:limit] // Take first 10 (already sorted newest first)
	}

	// Convert to JSON-friendly objects with ISO8601 UTC timestamps
	jsonHofs := make([]model.GlobalEntryJSON, len(hofs))
	for i, h := range hofs {
		jsonHofs[i] = model.GlobalEntryJSON{
			Timestamp:  h.Timestamp.UTC().Format(time.RFC3339),
			Type:       h.Type,
			PlayerName: h.PlayerName,
			TeamName:   h.TeamName,
			Target:     h.Target,
			Value:      h.Value,
			Location:   h.Location,
			IsHof:      h.IsHof,
			RawMessage: h.RawMessage,
		}
	}

	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(jsonHofs)
}

// handleWebSocket handles WebSocket connections
func (s *WebService) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Error("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()
	// Register client
	s.clientsLock.Lock()
	s.clients[conn] = true
	s.writeLockers[conn] = &sync.Mutex{}
	s.clientsLock.Unlock()
	// Remove client when connection closes
	defer func() {
		s.clientsLock.Lock()
		delete(s.clients, conn)
		delete(s.writeLockers, conn)
		s.clientsLock.Unlock()
	}()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// BroadcastEvent sends an event to all connected WebSocket clients
func (s *WebService) BroadcastEvent(eventType string, data interface{}) {
	// Format timestamps for global and hof events
	if eventType == "new_global" || eventType == "new_hof" {
		if entry, ok := data.(*storage.GlobalEntry); ok {
			// Convert to JSON-friendly object with ISO8601 UTC timestamp
			jsonEntry := model.GlobalEntryJSON{
				Timestamp:  entry.Timestamp.UTC().Format(time.RFC3339),
				Type:       entry.Type,
				PlayerName: entry.PlayerName,
				TeamName:   entry.TeamName,
				Target:     entry.Target,
				Value:      entry.Value,
				Location:   entry.Location,
				IsHof:      entry.IsHof,
				RawMessage: entry.RawMessage,
			}
			data = jsonEntry
		}
	}

	event := map[string]interface{}{
		"type": eventType,
		"data": data,
		"time": time.Now().UTC().Format(time.RFC3339),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		s.log.Error("Failed to marshal event: %v", err)
		return
	}
	s.clientsLock.Lock()
	clientsCopy := make(map[*websocket.Conn]*sync.Mutex)
	for client := range s.clients {
		clientsCopy[client] = s.writeLockers[client]
	}
	s.clientsLock.Unlock()

	for client, mutex := range clientsCopy {
		go func(c *websocket.Conn, mu *sync.Mutex) {
			mu.Lock()
			defer mu.Unlock()

			if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
				s.log.Error("Failed to send WebSocket message: %v", err)
				c.Close()

				s.clientsLock.Lock()
				delete(s.clients, c)
				delete(s.writeLockers, c)
				s.clientsLock.Unlock()
			}
		}(client, mutex)
	}
}
