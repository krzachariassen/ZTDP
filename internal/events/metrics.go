package events

import (
	"time"
)

// EventBusMetrics contains event bus performance metrics
type EventBusMetrics struct {
	TotalEvents     int64            `json:"total_events"`
	EventsPerSecond float64          `json:"events_per_second"`
	ErrorRate       float64          `json:"error_rate"`
	AvgLatency      time.Duration    `json:"avg_latency"`
	ActiveAgents    int              `json:"active_agents"`
	EventTypeStats  map[string]int64 `json:"event_type_stats"`
	LastUpdated     time.Time        `json:"last_updated"`
}

// MetricsCollector collects and aggregates event bus metrics
type MetricsCollector struct {
	startTime      time.Time
	totalEvents    int64
	errorCount     int64
	latencySum     time.Duration
	eventTypeCount map[string]int64
	activeAgents   map[string]bool
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime:      time.Now(),
		eventTypeCount: make(map[string]int64),
		activeAgents:   make(map[string]bool),
	}
}

// RecordEvent records an event for metrics
func (m *MetricsCollector) RecordEvent(eventType string, latency time.Duration, err error) {
	m.totalEvents++
	m.latencySum += latency
	m.eventTypeCount[eventType]++

	if err != nil {
		m.errorCount++
	}
}

// RecordAgentActivity records agent activity
func (m *MetricsCollector) RecordAgentActivity(agentID string) {
	m.activeAgents[agentID] = true
}

// GetMetrics returns current metrics snapshot
func (m *MetricsCollector) GetMetrics() EventBusMetrics {
	duration := time.Since(m.startTime)

	var avgLatency time.Duration
	if m.totalEvents > 0 {
		avgLatency = m.latencySum / time.Duration(m.totalEvents)
	}

	var eventsPerSecond float64
	if duration.Seconds() > 0 {
		eventsPerSecond = float64(m.totalEvents) / duration.Seconds()
	}

	var errorRate float64
	if m.totalEvents > 0 {
		errorRate = float64(m.errorCount) / float64(m.totalEvents)
	}

	eventTypeStats := make(map[string]int64)
	for k, v := range m.eventTypeCount {
		eventTypeStats[k] = v
	}

	return EventBusMetrics{
		TotalEvents:     m.totalEvents,
		EventsPerSecond: eventsPerSecond,
		ErrorRate:       errorRate,
		AvgLatency:      avgLatency,
		ActiveAgents:    len(m.activeAgents),
		EventTypeStats:  eventTypeStats,
		LastUpdated:     time.Now(),
	}
}

// ResetMetrics resets all metrics counters
func (m *MetricsCollector) ResetMetrics() {
	m.startTime = time.Now()
	m.totalEvents = 0
	m.errorCount = 0
	m.latencySum = 0
	m.eventTypeCount = make(map[string]int64)
	m.activeAgents = make(map[string]bool)
}
