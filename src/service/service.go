package service

// Service defines the interface for application services
type Service interface {
	// Initialize initializes the service
	Initialize() error
	
	// Name returns the name of the service
	Name() string
	
	// Run executes the service logic
	Run() error
	
	// Stop gracefully stops the service
	Stop() error
}

// BaseService provides a base implementation of the Service interface
type BaseService struct {
	name string
}

// NewBaseService creates a new BaseService
func NewBaseService(name string) *BaseService {
	return &BaseService{
		name: name,
	}
}

// Name returns the service name
func (s *BaseService) Name() string {
	return s.name
}

// Initialize provides a default implementation
func (s *BaseService) Initialize() error {
	return nil
}

// Run provides a default implementation
func (s *BaseService) Run() error {
	return nil
}

// Stop provides a default implementation
func (s *BaseService) Stop() error {
	return nil
}
