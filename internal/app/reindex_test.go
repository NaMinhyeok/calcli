package app

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
	"github.com/NaMinhyeok/calcli/internal/storage/cache"
)

func TestReindexHandler(t *testing.T) {
	t.Run("clear cache successfully", func(t *testing.T) {
		eventCache := cache.NewEventCache(100, true)

		// Add some events to cache
		event1 := domain.Event{UID: "event1", Summary: "Event 1"}
		event2 := domain.Event{UID: "event2", Summary: "Event 2"}
		eventCache.Set("home", "event1", event1, time.Now())
		eventCache.Set("home", "event2", event2, time.Now())

		if eventCache.Size() != 2 {
			t.Fatalf("Expected cache size 2, got %d", eventCache.Size())
		}

		var buf bytes.Buffer
		err := ReindexHandler(eventCache, &buf)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if eventCache.Size() != 0 {
			t.Errorf("Expected cache size 0 after reindex, got %d", eventCache.Size())
		}

		output := buf.String()
		if !strings.Contains(output, "Removed 2 cached events") {
			t.Errorf("Expected output to mention 2 events, got: %s", output)
		}
		if !strings.Contains(output, "Next read will rebuild") {
			t.Errorf("Expected output to mention rebuild, got: %s", output)
		}
	})

	t.Run("nil cache returns error", func(t *testing.T) {
		var buf bytes.Buffer
		err := ReindexHandler(nil, &buf)
		if err == nil {
			t.Error("Expected error for nil cache, got nil")
		}
		if !strings.Contains(err.Error(), "not enabled") {
			t.Errorf("Expected 'not enabled' error, got: %v", err)
		}
	})

	t.Run("empty cache", func(t *testing.T) {
		eventCache := cache.NewEventCache(100, true)

		var buf bytes.Buffer
		err := ReindexHandler(eventCache, &buf)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Removed 0 cached events") {
			t.Errorf("Expected output to mention 0 events, got: %s", output)
		}
	})
}
