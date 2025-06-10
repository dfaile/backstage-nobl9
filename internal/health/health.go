package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Component represents a system component that can be monitored
type Component struct {
	Name      string
	Status    Status
	LastCheck time.Time
	Error     error
}

// Checker defines the interface for health checks
type Checker interface {
	Check(ctx context.Context) error
}

// Health represents the health monitoring system
type Health struct {
	components map[string]*Component
	checkers   map[string]Checker
	mu         sync.RWMutex
	interval   time.Duration
}

// New creates a new health monitoring system
func New(interval time.Duration) *Health {
	return &Health{
		components: make(map[string]*Component),
		checkers:   make(map[string]Checker),
		interval:   interval,
	}
}

// Register adds a new component to monitor
func (h *Health) Register(name string, checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.components[name] = &Component{
		Name:      name,
		Status:    StatusUnhealthy,
		LastCheck: time.Time{},
	}
	h.checkers[name] = checker
}

// Start begins the health monitoring process
func (h *Health) Start(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.checkAll(ctx)
		}
	}
}

// checkAll performs health checks for all registered components
func (h *Health) checkAll(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for name, checker := range h.checkers {
		component := h.components[name]
		err := checker.Check(ctx)
		
		component.LastCheck = time.Now()
		if err != nil {
			component.Status = StatusUnhealthy
			component.Error = err
		} else {
			component.Status = StatusHealthy
			component.Error = nil
		}
	}
}

// GetStatus returns the current health status of all components
func (h *Health) GetStatus() map[string]*Component {
	h.mu.RLock()
	defer h.mu.RUnlock()

	status := make(map[string]*Component)
	for name, component := range h.components {
		status[name] = &Component{
			Name:      component.Name,
			Status:    component.Status,
			LastCheck: component.LastCheck,
			Error:     component.Error,
		}
	}
	return status
}

// IsHealthy returns true if all components are healthy
func (h *Health) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, component := range h.components {
		if component.Status != StatusHealthy {
			return false
		}
	}
	return true
}

// FormatStatus returns a formatted string of the health status
func (h *Health) FormatStatus() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var status strings.Builder
	status.WriteString("Health Status:\n")

	for _, component := range h.components {
		status.WriteString(fmt.Sprintf("- %s: %s", component.Name, component.Status))
		if component.Error != nil {
			status.WriteString(fmt.Sprintf(" (Error: %v)", component.Error))
		}
		status.WriteString(fmt.Sprintf(" (Last check: %s)\n", component.LastCheck.Format(time.RFC3339)))
	}

	return status.String()
} 