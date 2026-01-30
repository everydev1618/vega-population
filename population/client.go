package population

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DefaultSource is the default URL for the vega-population repository.
	DefaultSource = "https://raw.githubusercontent.com/martellcode/vega-population/main/"

	// DefaultCacheDir is the default cache directory relative to vega home.
	DefaultCacheDir = "cache/population"

	// DefaultVegaHome is the default vega home directory.
	DefaultVegaHome = ".vega"
)

// Client is the main entry point for library users.
type Client struct {
	source     string
	cacheDir   string
	installDir string
	noCache    bool
	cache      *Cache
}

// Option configures a Client.
type Option func(*Client)

// WithSource sets a custom source URL or local path.
func WithSource(url string) Option {
	return func(c *Client) {
		c.source = url
	}
}

// WithCacheDir sets a custom cache directory.
func WithCacheDir(path string) Option {
	return func(c *Client) {
		c.cacheDir = path
	}
}

// WithInstallDir sets a custom installation directory.
func WithInstallDir(path string) Option {
	return func(c *Client) {
		c.installDir = path
	}
}

// WithNoCache disables caching of index files.
func WithNoCache() Option {
	return func(c *Client) {
		c.noCache = true
	}
}

// NewClient creates a new population Client with the given options.
func NewClient(opts ...Option) (*Client, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	vegaHome := filepath.Join(home, DefaultVegaHome)

	c := &Client{
		source:     DefaultSource,
		cacheDir:   filepath.Join(vegaHome, DefaultCacheDir),
		installDir: vegaHome,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Initialize cache
	c.cache = NewCache(c.cacheDir, c.noCache)

	return c, nil
}

// Search returns matching items across all types.
func (c *Client) Search(ctx context.Context, query string, opts *SearchOptions) ([]SearchResult, error) {
	if opts == nil {
		opts = &SearchOptions{}
	}

	source := NewSource(c.source, c.cache)
	return source.Search(ctx, query, opts)
}

// Install installs an item by name.
// The name can be prefixed with @ for personas or + for profiles.
func (c *Client) Install(ctx context.Context, name string, opts *InstallOptions) error {
	if opts == nil {
		opts = &InstallOptions{}
	}

	kind, itemName := ParseItemName(name)
	source := NewSource(c.source, c.cache)

	return source.Install(ctx, kind, itemName, c.installDir, opts)
}

// List returns installed items of the given kind.
// If kind is empty, returns all installed items.
func (c *Client) List(kind ItemKind) ([]InstalledItem, error) {
	var items []InstalledItem

	kinds := []ItemKind{KindSkill, KindPersona, KindProfile}
	if kind != "" {
		kinds = []ItemKind{kind}
	}

	for _, k := range kinds {
		dir := filepath.Join(c.installDir, k.Plural())
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("reading %s directory: %w", k.Plural(), err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			manifestPath := filepath.Join(dir, entry.Name(), "vega.yaml")
			if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
				continue
			}

			manifest, err := LoadManifest(manifestPath)
			if err != nil {
				// Skip items with invalid manifests
				continue
			}

			items = append(items, InstalledItem{
				Kind:    k,
				Name:    entry.Name(),
				Version: manifest.Version,
				Path:    filepath.Join(dir, entry.Name()),
			})
		}
	}

	return items, nil
}

// Info returns detailed information about an item.
func (c *Client) Info(ctx context.Context, name string) (*ItemInfo, error) {
	kind, itemName := ParseItemName(name)
	source := NewSource(c.source, c.cache)

	return source.Info(ctx, kind, itemName, c.installDir)
}

// UpdateCache refreshes the cached index files.
func (c *Client) UpdateCache(ctx context.Context) error {
	source := NewSource(c.source, c.cache)
	return source.UpdateCache(ctx)
}

// Source returns the configured source URL.
func (c *Client) Source() string {
	return c.source
}

// InstallDir returns the configured installation directory.
func (c *Client) InstallDir() string {
	return c.installDir
}
