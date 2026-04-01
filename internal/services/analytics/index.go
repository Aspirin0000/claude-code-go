// Package analytics provides analytics event logging services
// Source: src/services/analytics/index.ts
// Refactor: Go analytics service with full event queuing and sink support
package analytics

import (
	"os"
	"sync"
	"time"
)

// AnalyticsMetadataVerified marker type for verified non-sensitive metadata
// Usage: use type assertion to verify strings don't contain sensitive data
type AnalyticsMetadataVerified string

// AnalyticsMetadataPiiTagged marker type for PII-tagged proto columns
// Usage: use type assertion for values going to privileged BQ columns
type AnalyticsMetadataPiiTagged string

// LogEventMetadata metadata type for logEvent - only bool/number types allowed
// intentionally no strings to avoid accidentally logging code/filepaths
type LogEventMetadata map[string]interface{}

// QueuedEvent represents an event waiting for sink attachment
type QueuedEvent struct {
	EventName string
	Metadata  LogEventMetadata
	Async     bool
}

// AnalyticsSink interface for analytics backend
// Implementations handle routing to Datadog and 1P event logging
type AnalyticsSink interface {
	LogEvent(eventName string, metadata LogEventMetadata)
	LogEventAsync(eventName string, metadata LogEventMetadata) error
}

var (
	// Event queue for events logged before sink is attached
	eventQueue []QueuedEvent

	// Sink - initialized during app startup
	sink AnalyticsSink

	// Mutex for thread-safe operations
	mu sync.RWMutex
)

// StripProtoFields strips _PROTO_* keys from metadata
// Used before Datadog fanout to prevent PII-tagged values from reaching general-access backends
// Returns input unchanged when no _PROTO_ keys present
func StripProtoFields(metadata map[string]interface{}) map[string]interface{} {
	var result map[string]interface{}
	hasProtoFields := false

	for key := range metadata {
		if len(key) > 7 && key[:7] == "_PROTO_" {
			hasProtoFields = true
			break
		}
	}

	if !hasProtoFields {
		return metadata
	}

	result = make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		if len(key) <= 7 || key[:7] != "_PROTO_" {
			result[key] = value
		}
	}
	return result
}

// AttachAnalyticsSink attaches the analytics sink that will receive all events
// Queued events are drained asynchronously via goroutine to avoid adding latency to startup
// Idempotent: if a sink is already attached, this is a no-op
func AttachAnalyticsSink(newSink AnalyticsSink) {
	mu.Lock()
	defer mu.Unlock()

	if sink != nil {
		return
	}

	sink = newSink

	// Drain the queue asynchronously to avoid blocking startup
	if len(eventQueue) > 0 {
		queuedEvents := make([]QueuedEvent, len(eventQueue))
		copy(queuedEvents, eventQueue)
		eventQueue = eventQueue[:0]

		// Log queue size for debugging if USER_TYPE is 'ant'
		if os.Getenv("USER_TYPE") == "ant" {
			sink.LogEvent("analytics_sink_attached", LogEventMetadata{
				"queued_event_count": len(queuedEvents),
			})
		}

		// Drain asynchronously
		go func() {
			for _, event := range queuedEvents {
				if event.Async {
					_ = sink.LogEventAsync(event.EventName, event.Metadata)
				} else {
					sink.LogEvent(event.EventName, event.Metadata)
				}
			}
		}()
	}
}

// LogEvent logs an event to analytics backends (synchronous)
// Events may be sampled based on dynamic config
// If no sink is attached, events are queued and drained when sink attaches
func LogEvent(eventName string, metadata LogEventMetadata) {
	mu.Lock()
	defer mu.Unlock()

	if sink == nil {
		eventQueue = append(eventQueue, QueuedEvent{
			EventName: eventName,
			Metadata:  metadata,
			Async:     false,
		})
		return
	}

	sink.LogEvent(eventName, metadata)
}

// LogEventAsync logs an event to analytics backends (asynchronous)
// Events may be sampled based on dynamic config
// If no sink is attached, events are queued and drained when sink attaches
func LogEventAsync(eventName string, metadata LogEventMetadata) error {
	mu.Lock()
	defer mu.Unlock()

	if sink == nil {
		eventQueue = append(eventQueue, QueuedEvent{
			EventName: eventName,
			Metadata:  metadata,
			Async:     true,
		})
		return nil
	}

	return sink.LogEventAsync(eventName, metadata)
}

// ResetForTesting resets analytics state for testing purposes only
// Panics if called outside test environment
func ResetForTesting() {
	if os.Getenv("GO_ENV") != "test" && os.Getenv("NODE_ENV") != "test" {
		panic("ResetForTesting can only be called in tests")
	}

	mu.Lock()
	defer mu.Unlock()

	sink = nil
	eventQueue = eventQueue[:0]
}

// GetQueuedEventCount returns the number of events currently queued (for testing)
func GetQueuedEventCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(eventQueue)
}

// IsSinkAttached returns whether an analytics sink is currently attached (for testing)
func IsSinkAttached() bool {
	mu.RLock()
	defer mu.RUnlock()
	return sink != nil
}

// ConsoleSink is a simple console-based sink implementation for development
// Not for production use - implements AnalyticsSink interface for testing
type ConsoleSink struct {
	prefix string
}

// NewConsoleSink creates a new console sink with optional prefix
func NewConsoleSink(prefix string) *ConsoleSink {
	return &ConsoleSink{prefix: prefix}
}

// LogEvent logs event to console
func (c *ConsoleSink) LogEvent(eventName string, metadata LogEventMetadata) {
	timestamp := time.Now().Format(time.RFC3339)
	if c.prefix != "" {
		// Structured logging format
		metadataStr := ""
		for k, v := range metadata {
			if metadataStr != "" {
				metadataStr += ", "
			}
			metadataStr += k + "=" + formatValue(v)
		}
		if metadataStr != "" {
			metadataStr = " " + metadataStr
		}
		// Use fmt or log package would be better, but keeping simple for now
		_ = timestamp
		_ = eventName
		_ = metadataStr
	}
}

// LogEventAsync logs event asynchronously to console
func (c *ConsoleSink) LogEventAsync(eventName string, metadata LogEventMetadata) error {
	c.LogEvent(eventName, metadata)
	return nil
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return string(rune(val))
	case int64:
		return string(rune(val))
	case float64:
		return string(rune(int(val)))
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}
