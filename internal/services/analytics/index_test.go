package analytics

import (
	"os"
	"testing"
)

func TestStripProtoFields(t *testing.T) {
	meta := map[string]interface{}{
		"foo":           "bar",
		"_PROTO_secret": "should_be_removed",
		"count":         42,
	}

	result := StripProtoFields(meta)
	if _, ok := result["_PROTO_secret"]; ok {
		t.Error("expected _PROTO_secret to be stripped")
	}
	if result["foo"] != "bar" {
		t.Error("expected foo to be preserved")
	}
	if result["count"] != 42 {
		t.Error("expected count to be preserved")
	}
}

func TestStripProtoFieldsNoProto(t *testing.T) {
	meta := map[string]interface{}{
		"foo": "bar",
	}
	result := StripProtoFields(meta)
	if len(result) != 1 {
		t.Errorf("expected length 1, got %d", len(result))
	}
}

func TestLogEventQueuesBeforeSink(t *testing.T) {
	os.Setenv("GO_ENV", "test")
	ResetForTesting()

	LogEvent("test_event", LogEventMetadata{"key": 1})
	if GetQueuedEventCount() != 1 {
		t.Fatalf("expected 1 queued event, got %d", GetQueuedEventCount())
	}

	sink := NewConsoleSink("")
	AttachAnalyticsSink(sink)

	// After attaching, queue should be drained
	if GetQueuedEventCount() != 0 {
		t.Errorf("expected queue to be drained, got %d", GetQueuedEventCount())
	}
	if !IsSinkAttached() {
		t.Error("expected sink to be attached")
	}
}

func TestAttachAnalyticsSinkIdempotent(t *testing.T) {
	os.Setenv("GO_ENV", "test")
	ResetForTesting()

	sink1 := NewConsoleSink("s1")
	sink2 := NewConsoleSink("s2")

	AttachAnalyticsSink(sink1)
	AttachAnalyticsSink(sink2)

	if !IsSinkAttached() {
		t.Fatal("expected sink to be attached")
	}
}

func TestLogEventAsyncQueuesBeforeSink(t *testing.T) {
	os.Setenv("GO_ENV", "test")
	ResetForTesting()

	_ = LogEventAsync("async_event", LogEventMetadata{"async": true})
	if GetQueuedEventCount() != 1 {
		t.Fatalf("expected 1 queued async event, got %d", GetQueuedEventCount())
	}
}

func TestConsoleSink(t *testing.T) {
	sink := NewConsoleSink("[Test]")
	// Just ensure it doesn't panic
	sink.LogEvent("event", LogEventMetadata{"a": 1, "b": "x", "c": true})
	_ = sink.LogEventAsync("async_event", LogEventMetadata{"d": 1.5})
}

func TestFormatValue(t *testing.T) {
	cases := []struct {
		input    interface{}
		expected string
	}{
		{"hello", "hello"},
		{42, "42"},
		{int64(99), "99"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
		{struct{}{}, ""},
	}

	for _, c := range cases {
		got := formatValue(c.input)
		if got != c.expected {
			t.Errorf("formatValue(%v) = %q, want %q", c.input, got, c.expected)
		}
	}
}

func TestResetForTestingPanicsOutsideTest(t *testing.T) {
	os.Unsetenv("GO_ENV")
	os.Unsetenv("NODE_ENV")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when calling ResetForTesting outside test env")
		}
	}()
	ResetForTesting()
}
