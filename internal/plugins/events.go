package plugins

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nix-ai-help/pkg/logger"
)

// EventBus handles plugin events and communication
type EventBus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
	logger   *logger.Logger
	buffer   *EventBuffer
}

// NewEventBus creates a new event bus
func NewEventBus(log *logger.Logger) *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
		logger:   log,
		buffer:   NewEventBuffer(1000), // Buffer last 1000 events
	}
}

// Subscribe adds an event handler for all events
func (eb *EventBus) Subscribe(handler EventHandler) error {
	return eb.SubscribeToType("*", handler)
}

// SubscribeToType adds an event handler for specific event types
func (eb *EventBus) SubscribeToType(eventType string, handler EventHandler) error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if eb.handlers[eventType] == nil {
		eb.handlers[eventType] = make([]EventHandler, 0)
	}

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	eb.logger.Debug(fmt.Sprintf("Subscribed handler to event type: %s", eventType))
	return nil
}

// Unsubscribe removes an event handler
func (eb *EventBus) Unsubscribe(handler EventHandler) error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	for eventType, handlers := range eb.handlers {
		for i, h := range handlers {
			// Compare function pointers (this is simplified)
			if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
				eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				eb.logger.Debug(fmt.Sprintf("Unsubscribed handler from event type: %s", eventType))
				return nil
			}
		}
	}

	return fmt.Errorf("handler not found")
}

// Emit publishes an event to all subscribers
func (eb *EventBus) Emit(event PluginEvent) {
	// Add to buffer
	eb.buffer.Add(event)

	// Set ID if not provided
	if event.ID == "" {
		event.ID = generateEventID()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	eb.logger.Debug(fmt.Sprintf("Emitting event: %s from %s", event.Type, event.Source))

	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	// Send to handlers subscribed to this event type
	if handlers, exists := eb.handlers[event.Type]; exists {
		eb.notifyHandlers(handlers, event)
	}

	// Send to handlers subscribed to all events
	if handlers, exists := eb.handlers["*"]; exists {
		eb.notifyHandlers(handlers, event)
	}
}

// notifyHandlers sends the event to a list of handlers
func (eb *EventBus) notifyHandlers(handlers []EventHandler, event PluginEvent) {
	for _, handler := range handlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					eb.logger.Error(fmt.Sprintf("Event handler panicked: %v", r))
				}
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := h(ctx, event); err != nil {
				eb.logger.Warn(fmt.Sprintf("Event handler error: %v", err))
			}
		}(handler)
	}
}

// GetEventHistory returns recent events from the buffer
func (eb *EventBus) GetEventHistory(limit int) []PluginEvent {
	return eb.buffer.GetLast(limit)
}

// GetEventsByType returns recent events of a specific type
func (eb *EventBus) GetEventsByType(eventType string, limit int) []PluginEvent {
	return eb.buffer.GetByType(eventType, limit)
}

// GetEventsBySource returns recent events from a specific source
func (eb *EventBus) GetEventsBySource(source string, limit int) []PluginEvent {
	return eb.buffer.GetBySource(source, limit)
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

// EventBuffer provides a circular buffer for storing events
type EventBuffer struct {
	events []PluginEvent
	size   int
	index  int
	mutex  sync.RWMutex
}

// NewEventBuffer creates a new event buffer
func NewEventBuffer(size int) *EventBuffer {
	return &EventBuffer{
		events: make([]PluginEvent, size),
		size:   size,
		index:  0,
	}
}

// Add adds an event to the buffer
func (eb *EventBuffer) Add(event PluginEvent) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	eb.events[eb.index] = event
	eb.index = (eb.index + 1) % eb.size
}

// GetLast returns the last n events
func (eb *EventBuffer) GetLast(n int) []PluginEvent {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	if n > eb.size {
		n = eb.size
	}

	result := make([]PluginEvent, 0, n)

	// Start from the most recent and go backwards
	for i := 0; i < n; i++ {
		idx := (eb.index - 1 - i + eb.size) % eb.size
		event := eb.events[idx]

		// Skip empty events (buffer not full yet)
		if event.ID == "" {
			break
		}

		result = append(result, event)
	}

	return result
}

// GetByType returns events of a specific type
func (eb *EventBuffer) GetByType(eventType string, limit int) []PluginEvent {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	result := make([]PluginEvent, 0)
	count := 0

	// Iterate through buffer from most recent
	for i := 0; i < eb.size && count < limit; i++ {
		idx := (eb.index - 1 - i + eb.size) % eb.size
		event := eb.events[idx]

		// Skip empty events
		if event.ID == "" {
			break
		}

		if event.Type == eventType {
			result = append(result, event)
			count++
		}
	}

	return result
}

// GetBySource returns events from a specific source
func (eb *EventBuffer) GetBySource(source string, limit int) []PluginEvent {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	result := make([]PluginEvent, 0)
	count := 0

	// Iterate through buffer from most recent
	for i := 0; i < eb.size && count < limit; i++ {
		idx := (eb.index - 1 - i + eb.size) % eb.size
		event := eb.events[idx]

		// Skip empty events
		if event.ID == "" {
			break
		}

		if event.Source == source {
			result = append(result, event)
			count++
		}
	}

	return result
}

// Clear clears all events from the buffer
func (eb *EventBuffer) Clear() {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	eb.events = make([]PluginEvent, eb.size)
	eb.index = 0
}

// MetricsCollector collects metrics about plugin operations
type MetricsCollector struct {
	metrics map[string]*PluginMetrics
	mutex   sync.RWMutex
	logger  *logger.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(log *logger.Logger) *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*PluginMetrics),
		logger:  log,
	}
}

// RecordExecution records an execution event for a plugin
func (mc *MetricsCollector) RecordExecution(pluginName string, duration time.Duration, success bool) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.metrics[pluginName] == nil {
		mc.metrics[pluginName] = &PluginMetrics{
			StartTime: time.Now(),
		}
	}

	metrics := mc.metrics[pluginName]
	metrics.ExecutionCount++
	metrics.TotalExecutionTime += duration
	metrics.AverageExecutionTime = time.Duration(int64(metrics.TotalExecutionTime) / metrics.ExecutionCount)
	metrics.LastExecutionTime = time.Now()

	if !success {
		metrics.ErrorCount++
	}

	if metrics.ExecutionCount > 0 {
		metrics.SuccessRate = float64(metrics.ExecutionCount-metrics.ErrorCount) / float64(metrics.ExecutionCount)
	}
}

// GetMetrics returns metrics for a plugin
func (mc *MetricsCollector) GetMetrics(pluginName string) (*PluginMetrics, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	metrics, exists := mc.metrics[pluginName]
	if !exists {
		return nil, false
	}

	// Return a copy
	metricsCopy := *metrics
	return &metricsCopy, true
}

// GetAllMetrics returns metrics for all plugins
func (mc *MetricsCollector) GetAllMetrics() map[string]*PluginMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	result := make(map[string]*PluginMetrics)
	for name, metrics := range mc.metrics {
		metricsCopy := *metrics
		result[name] = &metricsCopy
	}

	return result
}

// UpdateCustomMetric updates a custom metric for a plugin
func (mc *MetricsCollector) UpdateCustomMetric(pluginName, metricName string, value interface{}) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.metrics[pluginName] == nil {
		mc.metrics[pluginName] = &PluginMetrics{
			StartTime:     time.Now(),
			CustomMetrics: make(map[string]interface{}),
		}
	}

	if mc.metrics[pluginName].CustomMetrics == nil {
		mc.metrics[pluginName].CustomMetrics = make(map[string]interface{})
	}

	mc.metrics[pluginName].CustomMetrics[metricName] = value
}

// ResetMetrics resets metrics for a plugin
func (mc *MetricsCollector) ResetMetrics(pluginName string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	delete(mc.metrics, pluginName)
}

// EventListener provides a convenient way to listen to specific events
type EventListener struct {
	eventBus *EventBus
	filters  []EventFilter
	logger   *logger.Logger
}

// EventFilter defines criteria for filtering events
type EventFilter struct {
	Type   string
	Source string
	Tags   []string
}

// NewEventListener creates a new event listener
func NewEventListener(eventBus *EventBus, log *logger.Logger) *EventListener {
	return &EventListener{
		eventBus: eventBus,
		filters:  make([]EventFilter, 0),
		logger:   log,
	}
}

// AddFilter adds an event filter
func (el *EventListener) AddFilter(filter EventFilter) {
	el.filters = append(el.filters, filter)
}

// Listen starts listening for events that match the filters
func (el *EventListener) Listen(handler EventHandler) error {
	wrappedHandler := func(ctx context.Context, event PluginEvent) error {
		// Check if event matches any filter
		if len(el.filters) == 0 {
			// No filters, handle all events
			return handler(ctx, event)
		}

		for _, filter := range el.filters {
			if el.matchesFilter(event, filter) {
				return handler(ctx, event)
			}
		}

		// Event doesn't match any filter
		return nil
	}

	return el.eventBus.Subscribe(wrappedHandler)
}

// matchesFilter checks if an event matches a filter
func (el *EventListener) matchesFilter(event PluginEvent, filter EventFilter) bool {
	// Check type
	if filter.Type != "" && filter.Type != "*" && event.Type != filter.Type {
		return false
	}

	// Check source
	if filter.Source != "" && filter.Source != "*" && event.Source != filter.Source {
		return false
	}

	// Check tags
	if len(filter.Tags) > 0 {
		for _, filterTag := range filter.Tags {
			found := false
			for _, eventTag := range event.Tags {
				if eventTag == filterTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}
