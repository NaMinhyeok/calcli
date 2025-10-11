package app

import (
	"fmt"
	"io"

	"github.com/NaMinhyeok/calcli/internal/storage/cache"
)

// ReindexHandler clears the cache, forcing all events to be reloaded
func ReindexHandler(eventCache *cache.EventCache, output io.Writer) error {
	if eventCache == nil {
		return fmt.Errorf("cache is not enabled")
	}

	oldSize := eventCache.Size()
	eventCache.Clear()

	fmt.Fprintf(output, "Cache cleared. Removed %d cached events.\n", oldSize)
	fmt.Fprintf(output, "Next read will rebuild the cache.\n")

	return nil
}
