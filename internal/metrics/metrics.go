package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	TypeCounter   MetricType = "counter"
	TypeGauge     MetricType = "gauge"
	TypeHistogram MetricType = "histogram"
)

// Metric represents a single metric
type Metric struct {
	Name      string
	Type      MetricType
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
}

// Metrics represents the metrics collection system
type Metrics struct {
	metrics map[string]*Metric
	mu      sync.RWMutex
}

// New creates a new metrics collection system
func New() *Metrics {
	return &Metrics{
		metrics: make(map[string]*Metric),
	}
}

// Register registers a new metric
func (m *Metrics) Register(name string, metricType MetricType, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics[name] = &Metric{
		Name:      name,
		Type:      metricType,
		Value:     0,
		Labels:    labels,
		Timestamp: time.Now(),
	}
}

// Increment increments a counter metric
func (m *Metrics) Increment(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metric, exists := m.metrics[name]; exists && metric.Type == TypeCounter {
		metric.Value += value
		metric.Timestamp = time.Now()
	}
}

// Set sets a gauge metric
func (m *Metrics) Set(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metric, exists := m.metrics[name]; exists && metric.Type == TypeGauge {
		metric.Value = value
		metric.Timestamp = time.Now()
	}
}

// Observe records a value for a histogram metric
func (m *Metrics) Observe(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metric, exists := m.metrics[name]; exists && metric.Type == TypeHistogram {
		metric.Value = value
		metric.Timestamp = time.Now()
	}
}

// Get returns the current value of a metric
func (m *Metrics) Get(name string) (*Metric, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metric, exists := m.metrics[name]
	return metric, exists
}

// GetAll returns all metrics
func (m *Metrics) GetAll() map[string]*Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make(map[string]*Metric)
	for name, metric := range m.metrics {
		metrics[name] = &Metric{
			Name:      metric.Name,
			Type:      metric.Type,
			Value:     metric.Value,
			Labels:    metric.Labels,
			Timestamp: metric.Timestamp,
		}
	}
	return metrics
}

// FormatPrometheus returns metrics in Prometheus format
func (m *Metrics) FormatPrometheus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var output string
	for _, metric := range m.metrics {
		// Format labels
		var labels string
		if len(metric.Labels) > 0 {
			labels = "{"
			for k, v := range metric.Labels {
				labels += fmt.Sprintf("%s=%q,", k, v)
			}
			labels = labels[:len(labels)-1] + "}"
		}

		// Format metric
		output += fmt.Sprintf("%s%s %g %d\n",
			metric.Name,
			labels,
			metric.Value,
			metric.Timestamp.UnixNano()/1e9,
		)
	}
	return output
}

// StartCollector starts a background collector
func (m *Metrics) StartCollector(ctx context.Context, interval time.Duration, collector func() map[string]float64) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := collector()
			for name, value := range metrics {
				if metric, exists := m.metrics[name]; exists {
					switch metric.Type {
					case TypeCounter:
						m.Increment(name, value)
					case TypeGauge:
						m.Set(name, value)
					case TypeHistogram:
						m.Observe(name, value)
					}
				}
			}
		}
	}
} 