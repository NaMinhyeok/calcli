package cache

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

func TestEventCache_GetSet(t *testing.T) {
	cache := NewEventCache(100, true)

	event := domain.Event{
		UID:     "test-uid",
		Summary: "Test Event",
	}
	modTime := time.Now()

	// Test cache miss
	_, ok := cache.Get("home", "test-uid")
	if ok {
		t.Error("Expected cache miss, got hit")
	}

	// Test cache set and hit
	cache.Set("home", "test-uid", event, modTime)
	cached, ok := cache.Get("home", "test-uid")
	if !ok {
		t.Fatal("Expected cache hit, got miss")
	}

	if cached.Event.UID != event.UID {
		t.Errorf("Expected UID %s, got %s", event.UID, cached.Event.UID)
	}

	if !cached.ModTime.Equal(modTime) {
		t.Errorf("Expected modTime %v, got %v", modTime, cached.ModTime)
	}
}

func TestEventCache_Delete(t *testing.T) {
	cache := NewEventCache(100, true)

	event := domain.Event{
		UID:     "test-uid",
		Summary: "Test Event",
	}
	modTime := time.Now()

	cache.Set("home", "test-uid", event, modTime)

	// Verify it's in cache
	_, ok := cache.Get("home", "test-uid")
	if !ok {
		t.Fatal("Expected cache hit before delete")
	}

	// Delete and verify
	cache.Delete("home", "test-uid")
	_, ok = cache.Get("home", "test-uid")
	if ok {
		t.Error("Expected cache miss after delete")
	}

	// Verify size decreased
	if cache.Size() != 0 {
		t.Errorf("Expected size 0, got %d", cache.Size())
	}
}

func TestEventCache_Clear(t *testing.T) {
	cache := NewEventCache(100, true)

	// Add multiple events
	for i := 0; i < 5; i++ {
		event := domain.Event{
			UID:     string(rune('a' + i)),
			Summary: "Test Event",
		}
		cache.Set("home", event.UID, event, time.Now())
	}

	if cache.Size() != 5 {
		t.Errorf("Expected size 5, got %d", cache.Size())
	}

	// Clear cache
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	// Verify all entries are gone
	_, ok := cache.Get("home", "a")
	if ok {
		t.Error("Expected cache miss after clear")
	}
}

func TestEventCache_IsValid(t *testing.T) {
	cache := NewEventCache(100, true)

	event := domain.Event{
		UID:     "test-uid",
		Summary: "Test Event",
	}
	cachedTime := time.Now()

	cache.Set("home", "test-uid", event, cachedTime)

	tests := []struct {
		name        string
		fileModTime time.Time
		want        bool
	}{
		{
			name:        "file unchanged",
			fileModTime: cachedTime,
			want:        true,
		},
		{
			name:        "file older than cache",
			fileModTime: cachedTime.Add(-1 * time.Hour),
			want:        true,
		},
		{
			name:        "file newer than cache",
			fileModTime: cachedTime.Add(1 * time.Hour),
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cache.IsValid("home", "test-uid", tt.fileModTime)
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventCache_LoadOrStore(t *testing.T) {
	cache := NewEventCache(100, true)

	event := domain.Event{
		UID:     "test-uid",
		Summary: "Test Event",
	}

	loadCount := 0
	loader := func() (domain.Event, error) {
		loadCount++
		return event, nil
	}

	modTime := time.Now()

	// First load - should call loader
	result, err := cache.LoadOrStore("home", "test-uid", modTime, loader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.UID != event.UID {
		t.Errorf("Expected UID %s, got %s", event.UID, result.UID)
	}
	if loadCount != 1 {
		t.Errorf("Expected loader called once, got %d", loadCount)
	}

	// Second load - should use cache
	result, err = cache.LoadOrStore("home", "test-uid", modTime, loader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if loadCount != 1 {
		t.Errorf("Expected loader still called once (cached), got %d", loadCount)
	}

	// Load with newer file - should call loader again
	newerModTime := modTime.Add(1 * time.Hour)
	result, err = cache.LoadOrStore("home", "test-uid", newerModTime, loader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if loadCount != 2 {
		t.Errorf("Expected loader called twice (invalidated), got %d", loadCount)
	}
}

func TestEventCache_LoadOrStore_Error(t *testing.T) {
	cache := NewEventCache(100, true)

	expectedErr := errors.New("load error")
	loader := func() (domain.Event, error) {
		return domain.Event{}, expectedErr
	}

	_, err := cache.LoadOrStore("home", "test-uid", time.Now(), loader)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestEventCache_DisabledCache(t *testing.T) {
	cache := NewEventCache(100, false)

	event := domain.Event{
		UID:     "test-uid",
		Summary: "Test Event",
	}

	// Set should be no-op
	cache.Set("home", "test-uid", event, time.Now())

	// Get should always return miss
	_, ok := cache.Get("home", "test-uid")
	if ok {
		t.Error("Expected cache miss when disabled")
	}

	// LoadOrStore should always call loader
	loadCount := 0
	loader := func() (domain.Event, error) {
		loadCount++
		return event, nil
	}

	cache.LoadOrStore("home", "test-uid", time.Now(), loader)
	cache.LoadOrStore("home", "test-uid", time.Now(), loader)

	if loadCount != 2 {
		t.Errorf("Expected loader called twice when cache disabled, got %d", loadCount)
	}
}

func TestEventCache_Concurrency(t *testing.T) {
	cache := NewEventCache(1000, true)

	var wg sync.WaitGroup
	numGoroutines := 50
	numOpsPerGoroutine := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOpsPerGoroutine; j++ {
				event := domain.Event{
					UID:     string(rune('a' + (id*numOpsPerGoroutine+j)%26)),
					Summary: "Concurrent Event",
				}
				cache.Set("home", event.UID, event, time.Now())
			}
		}(i)
	}

	wg.Wait()

	// Verify cache integrity
	if cache.Size() < 1 {
		t.Error("Expected cache to contain events after concurrent writes")
	}
}

func TestEventCache_SingleflightPreventsDoubleLoad(t *testing.T) {
	cache := NewEventCache(100, true)

	loadCount := 0
	var mu sync.Mutex
	loader := func() (domain.Event, error) {
		mu.Lock()
		loadCount++
		mu.Unlock()
		time.Sleep(10 * time.Millisecond) // Simulate slow load
		return domain.Event{UID: "test-uid"}, nil
	}

	var wg sync.WaitGroup
	modTime := time.Now()

	// Start 10 concurrent loads for the same event
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.LoadOrStore("home", "test-uid", modTime, loader)
		}()
	}

	wg.Wait()

	// Singleflight should ensure loader is only called once
	if loadCount != 1 {
		t.Errorf("Expected loader called once due to singleflight, got %d", loadCount)
	}
}
