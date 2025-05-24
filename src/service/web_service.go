package service

import (
	"encoding/json"
	"eu-clams/internal/logger"
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
	upgrader    websocket.Upgrader
	templates   *template.Template
	templateDir string
	staticDir   string
}

// NewWebService creates a new WebService instance
func NewWebService(log *logger.Logger, db *storage.EntropyDB, playerName, teamName string, port int) *WebService {
	return &WebService{
		BaseService: NewBaseService("WebService"),
		log:         log,
		db:          db,
		playerName:  playerName,
		teamName:    teamName,
		port:        port,
		clients:     make(map[*websocket.Conn]bool),
		upgrader:    websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		templateDir: getTemplateDir(),
		staticDir:   getStaticDir(),
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

	// Generate stats
	statsData := s.db.GetStatsData()
	globals := s.db.GetPlayerGlobals()
	hofs := s.db.GetPlayerHofs()

	// Prepare template data
	data := map[string]interface{}{
		"PlayerName": s.playerName,
		"TeamName":   s.teamName,
		"Stats":      statsData,
		"Globals":    globals,
		"Hofs":       hofs,
		"Generated":  time.Now().Format("2006-01-02 15:04:05"),
	}

	// Render the template
	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		s.log.Error("Failed to render template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleStats handles the stats API endpoint
func (s *WebService) handleStats(w http.ResponseWriter, r *http.Request) {
	statsData := s.db.GetStatsData()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statsData)
}

// handleGlobals handles the globals API endpoint
func (s *WebService) handleGlobals(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	globals := s.db.GetPlayerGlobals()
	if len(globals) > limit {
		globals = globals[len(globals)-limit:]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(globals)
}

// handleHofs handles the HOFs API endpoint
func (s *WebService) handleHofs(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	hofs := s.db.GetPlayerHofs()
	if len(hofs) > limit {
		hofs = hofs[len(hofs)-limit:]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hofs)
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
	s.clientsLock.Unlock()

	// Remove client when connection closes
	defer func() {
		s.clientsLock.Lock()
		delete(s.clients, conn)
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
	event := map[string]interface{}{
		"type": eventType,
		"data": data,
		"time": time.Now().Format("2006-01-02 15:04:05"),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		s.log.Error("Failed to marshal event: %v", err)
		return
	}

	s.clientsLock.Lock()
	for client := range s.clients {
		go func(c *websocket.Conn) {
			if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
				s.log.Error("Failed to send WebSocket message: %v", err)
				c.Close()
				s.clientsLock.Lock()
				delete(s.clients, c)
				s.clientsLock.Unlock()
			}
		}(client)
	}
	s.clientsLock.Unlock()
}
