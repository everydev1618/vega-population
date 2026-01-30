package population

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// CacheTTL is the default cache time-to-live for index files.
	CacheTTL = 1 * time.Hour
)

// Cache handles local caching of index files.
type Cache struct {
	dir      string
	disabled bool
	ttl      time.Duration
}

// NewCache creates a new Cache instance.
func NewCache(dir string, disabled bool) *Cache {
	return &Cache{
		dir:      dir,
		disabled: disabled,
		ttl:      CacheTTL,
	}
}

// Get retrieves a cached file if it exists and is not expired.
// Returns the content and true if the cache is valid, nil and false otherwise.
func (c *Cache) Get(name string) ([]byte, bool) {
	if c.disabled {
		return nil, false
	}

	path := filepath.Join(c.dir, name)
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}

	// Check if cache is expired
	if time.Since(info.ModTime()) > c.ttl {
		return nil, false
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	return content, true
}

// Set stores content in the cache.
func (c *Cache) Set(name string, content []byte) error {
	if c.disabled {
		return nil
	}

	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	path := filepath.Join(c.dir, name)
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	return nil
}

// Invalidate removes a cached file.
func (c *Cache) Invalidate(name string) error {
	path := filepath.Join(c.dir, name)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing cache file: %w", err)
	}
	return nil
}

// InvalidateAll removes all cached files.
func (c *Cache) InvalidateAll() error {
	if err := os.RemoveAll(c.dir); err != nil {
		return fmt.Errorf("removing cache directory: %w", err)
	}
	return nil
}

// Dir returns the cache directory path.
func (c *Cache) Dir() string {
	return c.dir
}
