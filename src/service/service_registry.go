package service

import (
	"fmt"
	"sync"
)

// ServiceRegistry provides a centralized registry for keeping track of services
// This is useful for broadcasting events across services
type ServiceRegistry struct {
	webServices map[string]*WebService
	mu          sync.RWMutex
}

var (
	// Global service registry instance
	registry = &ServiceRegistry{
		webServices: make(map[string]*WebService),
	}
)

// RegisterWebService adds a web service to the registry
func RegisterWebService(name string, service *WebService) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.webServices[name] = service
}

// UnregisterWebService removes a web service from the registry
func UnregisterWebService(name string) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	delete(registry.webServices, name)
}

// BroadcastToWebServices sends an event to all registered web services
func BroadcastToWebServices(eventType string, data interface{}) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	for _, service := range registry.webServices {
		go service.BroadcastEvent(eventType, data)
	}
}

// GetWebServiceByPort returns a registered web service by its port number
func GetWebServiceByPort(port int) *WebService {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	serviceName := fmt.Sprintf("web_%d", port)
	return registry.webServices[serviceName]
}
