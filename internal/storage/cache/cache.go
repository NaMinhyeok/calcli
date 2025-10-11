package cache

import (
	"sync"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
	"golang.org/x/sync/singleflight"
)

// CachedEvent represents a cached event with its modification time
type CachedEvent struct {
	Event   domain.Event
	ModTime time.Time
}

// EventCache provides thread-safe caching for calendar events
type EventCache struct {
	cache   sync.Map           // key: "calendar/uid" -> *CachedEvent
	group   singleflight.Group // prevents duplicate loads
	maxSize int                // maximum cache size (0 = unlimited)
	mu      sync.RWMutex       // protects size counter
	size    int                // current cache size
	enabled bool               // cache enable/disable flag
}

// NewEventCache creates a new EventCache
func NewEventCache(maxSize int, enabled bool) *EventCache {
	return &EventCache{
		maxSize: maxSize,
		enabled: enabled,
	}
}

// Get retrieves an event from the cache
func (c *EventCache) Get(calendar, uid string) (*CachedEvent, bool) {
	if !c.enabled {
		return nil, false
	}

	key := makeKey(calendar, uid)
	val, ok := c.cache.Load(key)
	if !ok {
		return nil, false
	}

	return val.(*CachedEvent), true
}

// Set stores an event in the cache with its modification time
func (c *EventCache) Set(calendar, uid string, event domain.Event, modTime time.Time) {
	if !c.enabled {
		return
	}

	key := makeKey(calendar, uid)
	cached := &CachedEvent{
		Event:   event,
		ModTime: modTime,
	}

	// Check if this is a new entry
	_, loaded := c.cache.LoadOrStore(key, cached)

	if !loaded {
		// New entry - increment size
		c.mu.Lock()
		c.size++
		// If cache is full and maxSize > 0, we should evict (future enhancement)
		c.mu.Unlock()
	} else {
		// Update existing entry
		c.cache.Store(key, cached)
	}
}

// Delete removes an event from the cache
func (c *EventCache) Delete(calendar, uid string) {
	if !c.enabled {
		return
	}

	key := makeKey(calendar, uid)
	_, existed := c.cache.LoadAndDelete(key)

	if existed {
		c.mu.Lock()
		c.size--
		c.mu.Unlock()
	}
}

// Clear removes all entries from the cache
func (c *EventCache) Clear() {
	c.cache.Range(func(key, value interface{}) bool {
		c.cache.Delete(key)
		return true
	})

	c.mu.Lock()
	c.size = 0
	c.mu.Unlock()
}

// Size returns the current number of cached events
func (c *EventCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.size
}

// IsValid checks if a cached event is still valid based on file modification time
func (c *EventCache) IsValid(calendar, uid string, fileModTime time.Time) bool {
	if !c.enabled {
		return false
	}

	cached, ok := c.Get(calendar, uid)
	if !ok {
		return false
	}

	// Cache is valid if file hasn't been modified since caching
	return !fileModTime.After(cached.ModTime)
}

// LoadOrStore attempts to get from cache or executes the loader function
// Uses singleflight to prevent duplicate loads
func (c *EventCache) LoadOrStore(calendar, uid string, fileModTime time.Time, loader func() (domain.Event, error)) (domain.Event, error) {
	if !c.enabled {
		return loader()
	}

	// Check if cache is valid
	if c.IsValid(calendar, uid, fileModTime) {
		cached, _ := c.Get(calendar, uid)
		return cached.Event, nil
	}

	// Use singleflight to prevent duplicate loads
	key := makeKey(calendar, uid)
	val, err, _ := c.group.Do(key, func() (interface{}, error) {
		event, err := loader()
		if err != nil {
			return nil, err
		}

		// Cache the loaded event
		c.Set(calendar, uid, event, fileModTime)
		return event, nil
	})

	if err != nil {
		return domain.Event{}, err
	}

	return val.(domain.Event), nil
}

// makeKey creates a cache key from calendar and uid
func makeKey(calendar, uid string) string {
	return calendar + "/" + uid
}
