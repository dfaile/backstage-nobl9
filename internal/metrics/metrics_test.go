package metrics_test

import (
	"context"
	"testing"
	"time"
	"github.com/dfaile/backstage-nobl9/internal/metrics"
)

func TestNew(t *testing.T) {
	m := New()
	if m.metrics == nil {
		t.Error("expected metrics map to be initialized")
	}
}

func TestRegister(t *testing.T) {
	m := New()
	labels := map[string]string{"test": "label"}

	m.Register("test_counter", TypeCounter, labels)
	m.Register("test_gauge", TypeGauge, labels)
	m.Register("test_histogram", TypeHistogram, labels)

	if len(m.metrics) != 3 {
		t.Errorf("expected 3 metrics, got %d", len(m.metrics))
	}

	// Verify counter
	if metric, exists := m.metrics["test_counter"]; !exists {
		t.Error("expected counter metric to exist")
	} else {
		if metric.Type != TypeCounter {
			t.Errorf("expected type %s, got %s", TypeCounter, metric.Type)
		}
		if metric.Labels["test"] != "label" {
			t.Error("expected labels to be preserved")
		}
	}

	// Verify gauge
	if metric, exists := m.metrics["test_gauge"]; !exists {
		t.Error("expected gauge metric to exist")
	} else {
		if metric.Type != TypeGauge {
			t.Errorf("expected type %s, got %s", TypeGauge, metric.Type)
		}
	}

	// Verify histogram
	if metric, exists := m.metrics["test_histogram"]; !exists {
		t.Error("expected histogram metric to exist")
	} else {
		if metric.Type != TypeHistogram {
			t.Errorf("expected type %s, got %s", TypeHistogram, metric.Type)
		}
	}
}

func TestIncrement(t *testing.T) {
	m := New()
	m.Register("test_counter", TypeCounter, nil)

	// Test increment
	m.Increment("test_counter", 1.0)
	if metric, _ := m.Get("test_counter"); metric.Value != 1.0 {
		t.Errorf("expected value 1.0, got %f", metric.Value)
	}

	// Test multiple increments
	m.Increment("test_counter", 2.0)
	if metric, _ := m.Get("test_counter"); metric.Value != 3.0 {
		t.Errorf("expected value 3.0, got %f", metric.Value)
	}

	// Test non-existent metric
	m.Increment("non_existent", 1.0)
	if metric, exists := m.Get("non_existent"); exists {
		t.Error("expected metric to not exist")
	}

	// Test wrong type
	m.Register("test_gauge", TypeGauge, nil)
	m.Increment("test_gauge", 1.0)
	if metric, _ := m.Get("test_gauge"); metric.Value != 0.0 {
		t.Error("expected gauge value to remain unchanged")
	}
}

func TestSet(t *testing.T) {
	m := New()
	m.Register("test_gauge", TypeGauge, nil)

	// Test set
	m.Set("test_gauge", 1.0)
	if metric, _ := m.Get("test_gauge"); metric.Value != 1.0 {
		t.Errorf("expected value 1.0, got %f", metric.Value)
	}

	// Test update
	m.Set("test_gauge", 2.0)
	if metric, _ := m.Get("test_gauge"); metric.Value != 2.0 {
		t.Errorf("expected value 2.0, got %f", metric.Value)
	}

	// Test non-existent metric
	m.Set("non_existent", 1.0)
	if metric, exists := m.Get("non_existent"); exists {
		t.Error("expected metric to not exist")
	}

	// Test wrong type
	m.Register("test_counter", TypeCounter, nil)
	m.Set("test_counter", 1.0)
	if metric, _ := m.Get("test_counter"); metric.Value != 0.0 {
		t.Error("expected counter value to remain unchanged")
	}
}

func TestObserve(t *testing.T) {
	m := New()
	m.Register("test_histogram", TypeHistogram, nil)

	// Test observe
	m.Observe("test_histogram", 1.0)
	if metric, _ := m.Get("test_histogram"); metric.Value != 1.0 {
		t.Errorf("expected value 1.0, got %f", metric.Value)
	}

	// Test update
	m.Observe("test_histogram", 2.0)
	if metric, _ := m.Get("test_histogram"); metric.Value != 2.0 {
		t.Errorf("expected value 2.0, got %f", metric.Value)
	}

	// Test non-existent metric
	m.Observe("non_existent", 1.0)
	if metric, exists := m.Get("non_existent"); exists {
		t.Error("expected metric to not exist")
	}

	// Test wrong type
	m.Register("test_counter", TypeCounter, nil)
	m.Observe("test_counter", 1.0)
	if metric, _ := m.Get("test_counter"); metric.Value != 0.0 {
		t.Error("expected counter value to remain unchanged")
	}
}

func TestGetAll(t *testing.T) {
	m := New()
	m.Register("test1", TypeCounter, nil)
	m.Register("test2", TypeGauge, nil)

	metrics := m.GetAll()
	if len(metrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(metrics))
	}

	if _, exists := metrics["test1"]; !exists {
		t.Error("expected test1 metric to exist")
	}

	if _, exists := metrics["test2"]; !exists {
		t.Error("expected test2 metric to exist")
	}
}

func TestFormatPrometheus(t *testing.T) {
	m := New()
	labels := map[string]string{"test": "label"}
	m.Register("test_metric", TypeCounter, labels)
	m.Increment("test_metric", 1.0)

	output := m.FormatPrometheus()
	expected := `test_metric{test="label"} 1 `
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %s, got %s", expected, output)
	}
}

func TestStartCollector(t *testing.T) {
	m := New()
	m.Register("test_metric", TypeCounter, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	collector := func() map[string]float64 {
		return map[string]float64{
			"test_metric": 1.0,
		}
	}

	m.StartCollector(ctx, 10*time.Millisecond, collector)
	time.Sleep(20 * time.Millisecond)

	if metric, _ := m.Get("test_metric"); metric.Value == 0.0 {
		t.Error("expected metric to be updated by collector")
	}
} 