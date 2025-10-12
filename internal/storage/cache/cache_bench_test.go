package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

func BenchmarkEventCache_Set(b *testing.B) {
	cache := NewEventCache(10000, true)
	event := domain.Event{
		UID:     "test-uid",
		Summary: "Benchmark Event",
	}
	modTime := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("home", fmt.Sprintf("uid-%d", i), event, modTime)
	}
}

func BenchmarkEventCache_Get(b *testing.B) {
	cache := NewEventCache(10000, true)
	event := domain.Event{
		UID:     "test-uid",
		Summary: "Benchmark Event",
	}
	modTime := time.Now()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Set("home", fmt.Sprintf("uid-%d", i), event, modTime)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("home", fmt.Sprintf("uid-%d", i%1000))
	}
}

func BenchmarkEventCache_LoadOrStore(b *testing.B) {
	cache := NewEventCache(10000, true)
	event := domain.Event{
		UID:     "test-uid",
		Summary: "Benchmark Event",
	}
	modTime := time.Now()

	loader := func() (domain.Event, error) {
		return event, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.LoadOrStore("home", fmt.Sprintf("uid-%d", i%100), modTime, loader)
	}
}

func BenchmarkEventCache_ConcurrentSet(b *testing.B) {
	cache := NewEventCache(10000, true)
	event := domain.Event{
		UID:     "test-uid",
		Summary: "Benchmark Event",
	}
	modTime := time.Now()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Set("home", fmt.Sprintf("uid-%d", i), event, modTime)
			i++
		}
	})
}

func BenchmarkEventCache_ConcurrentGet(b *testing.B) {
	cache := NewEventCache(10000, true)
	event := domain.Event{
		UID:     "test-uid",
		Summary: "Benchmark Event",
	}
	modTime := time.Now()

	// Pre-populate cache
	for i := 0; i < 10000; i++ {
		cache.Set("home", fmt.Sprintf("uid-%d", i), event, modTime)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Get("home", fmt.Sprintf("uid-%d", i%10000))
			i++
		}
	})
}

func BenchmarkEventCache_ConcurrentLoadOrStore(b *testing.B) {
	cache := NewEventCache(10000, true)
	event := domain.Event{
		UID:     "test-uid",
		Summary: "Benchmark Event",
	}
	modTime := time.Now()

	loader := func() (domain.Event, error) {
		// Simulate small load time
		time.Sleep(10 * time.Microsecond)
		return event, nil
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Multiple goroutines trying to load same keys to test singleflight
			cache.LoadOrStore("home", fmt.Sprintf("uid-%d", i%10), modTime, loader)
			i++
		}
	})
}

func BenchmarkEventCache_MixedOperations(b *testing.B) {
	cache := NewEventCache(10000, true)
	event := domain.Event{
		UID:     "test-uid",
		Summary: "Benchmark Event",
	}
	modTime := time.Now()

	// Pre-populate cache
	for i := 0; i < 5000; i++ {
		cache.Set("home", fmt.Sprintf("uid-%d", i), event, modTime)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 3 {
			case 0:
				cache.Get("home", fmt.Sprintf("uid-%d", i%10000))
			case 1:
				cache.Set("home", fmt.Sprintf("uid-%d", i), event, modTime)
			case 2:
				cache.IsValid("home", fmt.Sprintf("uid-%d", i%5000), modTime)
			}
			i++
		}
	})
}

// Benchmark cache vs no-cache for typical workload
func BenchmarkEventCache_CacheVsNoCache(b *testing.B) {
	event := domain.Event{
		UID:     "test-uid",
		Summary: "Benchmark Event",
	}
	modTime := time.Now()

	loadCount := 0
	var mu sync.Mutex

	loader := func() (domain.Event, error) {
		mu.Lock()
		loadCount++
		mu.Unlock()
		// Simulate file I/O
		time.Sleep(100 * time.Microsecond)
		return event, nil
	}

	b.Run("WithCache", func(b *testing.B) {
		cache := NewEventCache(1000, true)
		loadCount = 0

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Repeatedly access same 100 events
			cache.LoadOrStore("home", fmt.Sprintf("uid-%d", i%100), modTime, loader)
		}

		b.ReportMetric(float64(loadCount), "loads")
	})

	b.Run("WithoutCache", func(b *testing.B) {
		cache := NewEventCache(1000, false)
		loadCount = 0

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Repeatedly access same 100 events
			cache.LoadOrStore("home", fmt.Sprintf("uid-%d", i%100), modTime, loader)
		}

		b.ReportMetric(float64(loadCount), "loads")
	})
}

func BenchmarkEventCache_LargeCache(b *testing.B) {
	sizes := []int{1000, 5000, 10000, 50000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			cache := NewEventCache(size, true)
			event := domain.Event{
				UID:     "test-uid",
				Summary: "Benchmark Event",
			}
			modTime := time.Now()

			// Pre-populate cache to size
			for i := 0; i < size; i++ {
				cache.Set("home", fmt.Sprintf("uid-%d", i), event, modTime)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cache.Get("home", fmt.Sprintf("uid-%d", i%size))
			}
		})
	}
}
