package health

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// mockChecker implements the Checker interface for testing
type mockChecker struct {
	err error
}

func (m *mockChecker) Check(ctx context.Context) error {
	return m.err
}

func TestNew(t *testing.T) {
	interval := 5 * time.Second
	h := New(interval)

	if h.interval != interval {
		t.Errorf("expected interval %v, got %v", interval, h.interval)
	}

	if h.components == nil {
		t.Error("expected components map to be initialized")
	}

	if h.checkers == nil {
		t.Error("expected checkers map to be initialized")
	}
}

func TestRegister(t *testing.T) {
	h := New(5 * time.Second)
	checker := &mockChecker{}

	h.Register("test", checker)

	if _, exists := h.components["test"]; !exists {
		t.Error("expected component to be registered")
	}

	if _, exists := h.checkers["test"]; !exists {
		t.Error("expected checker to be registered")
	}

	component := h.components["test"]
	if component.Status != StatusUnhealthy {
		t.Errorf("expected initial status %s, got %s", StatusUnhealthy, component.Status)
	}
}

func TestCheckAll(t *testing.T) {
	h := New(5 * time.Second)
	ctx := context.Background()

	// Test healthy component
	healthyChecker := &mockChecker{err: nil}
	h.Register("healthy", healthyChecker)

	// Test unhealthy component
	unhealthyChecker := &mockChecker{err: errors.New("test error")}
	h.Register("unhealthy", unhealthyChecker)

	h.checkAll(ctx)

	healthyComponent := h.components["healthy"]
	if healthyComponent.Status != StatusHealthy {
		t.Errorf("expected status %s, got %s", StatusHealthy, healthyComponent.Status)
	}

	unhealthyComponent := h.components["unhealthy"]
	if unhealthyComponent.Status != StatusUnhealthy {
		t.Errorf("expected status %s, got %s", StatusUnhealthy, unhealthyComponent.Status)
	}
	if unhealthyComponent.Error == nil {
		t.Error("expected error to be set")
	}
}

func TestGetStatus(t *testing.T) {
	h := New(5 * time.Second)
	checker := &mockChecker{err: nil}
	h.Register("test", checker)

	status := h.GetStatus()
	if len(status) != 1 {
		t.Errorf("expected 1 component, got %d", len(status))
	}

	component, exists := status["test"]
	if !exists {
		t.Error("expected component to exist in status")
	}

	if component.Name != "test" {
		t.Errorf("expected name 'test', got %s", component.Name)
	}
}

func TestIsHealthy(t *testing.T) {
	h := New(5 * time.Second)
	ctx := context.Background()

	// Test all healthy
	healthyChecker := &mockChecker{err: nil}
	h.Register("healthy1", healthyChecker)
	h.Register("healthy2", healthyChecker)
	h.checkAll(ctx)

	if !h.IsHealthy() {
		t.Error("expected system to be healthy")
	}

	// Test one unhealthy
	unhealthyChecker := &mockChecker{err: errors.New("test error")}
	h.Register("unhealthy", unhealthyChecker)
	h.checkAll(ctx)

	if h.IsHealthy() {
		t.Error("expected system to be unhealthy")
	}
}

func TestFormatStatus(t *testing.T) {
	h := New(5 * time.Second)
	ctx := context.Background()

	healthyChecker := &mockChecker{err: nil}
	h.Register("healthy", healthyChecker)

	unhealthyChecker := &mockChecker{err: errors.New("test error")}
	h.Register("unhealthy", unhealthyChecker)

	h.checkAll(ctx)
	status := h.FormatStatus()

	if status == "" {
		t.Error("expected non-empty status string")
	}

	// Verify status contains component names
	if !strings.Contains(status, "healthy") || !strings.Contains(status, "unhealthy") {
		t.Error("expected status to contain component names")
	}

	// Verify status contains error message
	if !strings.Contains(status, "test error") {
		t.Error("expected status to contain error message")
	}
} 