package population

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Source handles fetching content from local or remote sources.
type Source struct {
	baseURL string
	cache   *Cache
	isLocal bool
}

// NewSource creates a new Source instance.
func NewSource(baseURL string, cache *Cache) *Source {
	// Normalize the URL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	isLocal := !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://")

	return &Source{
		baseURL: baseURL,
		cache:   cache,
		isLocal: isLocal,
	}
}

// fetch retrieves content from the source.
func (s *Source) fetch(ctx context.Context, path string) ([]byte, error) {
	if s.isLocal {
		return s.fetchLocal(path)
	}
	return s.fetchRemote(ctx, path)
}

func (s *Source) fetchLocal(path string) ([]byte, error) {
	fullPath := filepath.Join(strings.TrimSuffix(s.baseURL, "/"), path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("reading local file %s: %w", fullPath, err)
	}
	return content, nil
}

func (s *Source) fetchRemote(ctx context.Context, path string) ([]byte, error) {
	url := s.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: status %d", url, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return content, nil
}

// Index file structures

// SkillsIndex represents the skills/index.yaml structure.
type SkillsIndex struct {
	Skills map[string]IndexEntry `yaml:"skills"`
}

// PersonasIndex represents the personas/index.yaml structure.
type PersonasIndex struct {
	Personas map[string]IndexEntry `yaml:"personas"`
}

// ProfilesIndex represents the profiles/index.yaml structure.
type ProfilesIndex struct {
	Profiles map[string]ProfileIndexEntry `yaml:"profiles"`
}

// IndexEntry represents an entry in the skills or personas index.
type IndexEntry struct {
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Tags        []string `yaml:"tags"`
	Tools       []string `yaml:"tools,omitempty"`
}

// ProfileIndexEntry represents an entry in the profiles index.
type ProfileIndexEntry struct {
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Persona     string   `yaml:"persona"`
	Skills      []string `yaml:"skills"`
}

// Manifest represents a vega.yaml file.
type Manifest struct {
	Kind              string   `yaml:"kind"`
	Name              string   `yaml:"name"`
	Version           string   `yaml:"version"`
	Description       string   `yaml:"description"`
	Author            string   `yaml:"author"`
	Tags              []string `yaml:"tags,omitempty"`
	Persona           string   `yaml:"persona,omitempty"`
	Skills            []string `yaml:"skills,omitempty"`
	RecommendedSkills []string `yaml:"recommended_skills,omitempty"`
	SystemPrompt      string   `yaml:"system_prompt,omitempty"`
}

// getIndex fetches and parses an index file.
func (s *Source) getIndex(ctx context.Context, kind ItemKind) (map[string]IndexEntry, map[string]ProfileIndexEntry, error) {
	indexPath := kind.Plural() + "/index.yaml"
	cacheKey := kind.Plural() + "-index.yaml"

	// Try cache first
	if content, ok := s.cache.Get(cacheKey); ok {
		return s.parseIndex(content, kind)
	}

	// Fetch from source
	content, err := s.fetch(ctx, indexPath)
	if err != nil {
		return nil, nil, err
	}

	// Cache the result
	if err := s.cache.Set(cacheKey, content); err != nil {
		// Log but don't fail on cache errors
		fmt.Fprintf(os.Stderr, "Warning: failed to cache %s: %v\n", cacheKey, err)
	}

	return s.parseIndex(content, kind)
}

func (s *Source) parseIndex(content []byte, kind ItemKind) (map[string]IndexEntry, map[string]ProfileIndexEntry, error) {
	switch kind {
	case KindSkill:
		var idx SkillsIndex
		if err := yaml.Unmarshal(content, &idx); err != nil {
			return nil, nil, fmt.Errorf("parsing skills index: %w", err)
		}
		return idx.Skills, nil, nil

	case KindPersona:
		var idx PersonasIndex
		if err := yaml.Unmarshal(content, &idx); err != nil {
			return nil, nil, fmt.Errorf("parsing personas index: %w", err)
		}
		return idx.Personas, nil, nil

	case KindProfile:
		var idx ProfilesIndex
		if err := yaml.Unmarshal(content, &idx); err != nil {
			return nil, nil, fmt.Errorf("parsing profiles index: %w", err)
		}
		return nil, idx.Profiles, nil

	default:
		return nil, nil, fmt.Errorf("unknown item kind: %s", kind)
	}
}

// GetManifest fetches a manifest file for a specific item.
func (s *Source) GetManifest(ctx context.Context, kind ItemKind, name string) (*Manifest, error) {
	path := fmt.Sprintf("%s/%s/vega.yaml", kind.Plural(), name)

	content, err := s.fetch(ctx, path)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}

	return &manifest, nil
}

// GetManifestRaw fetches the raw content of a manifest file.
func (s *Source) GetManifestRaw(ctx context.Context, kind ItemKind, name string) ([]byte, error) {
	path := fmt.Sprintf("%s/%s/vega.yaml", kind.Plural(), name)
	return s.fetch(ctx, path)
}

// LoadManifest loads a manifest from a local file path.
func LoadManifest(path string) (*Manifest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}

	return &manifest, nil
}

// Info returns detailed information about an item.
func (s *Source) Info(ctx context.Context, kind ItemKind, name string, installDir string) (*ItemInfo, error) {
	// Fetch from index first for basic info
	entries, profiles, err := s.getIndex(ctx, kind)
	if err != nil {
		return nil, err
	}

	info := &ItemInfo{
		Kind: kind,
		Name: name,
	}

	if kind == KindProfile {
		entry, ok := profiles[name]
		if !ok {
			return nil, fmt.Errorf("%s %q not found", kind, name)
		}
		info.Version = entry.Version
		info.Description = entry.Description
		info.Author = entry.Author
		info.Persona = entry.Persona
		info.Skills = entry.Skills
	} else {
		entry, ok := entries[name]
		if !ok {
			return nil, fmt.Errorf("%s %q not found", kind, name)
		}
		info.Version = entry.Version
		info.Description = entry.Description
		info.Author = entry.Author
		info.Tags = entry.Tags
	}

	// Check if installed
	installedPath := filepath.Join(installDir, kind.Plural(), name, "vega.yaml")
	if _, err := os.Stat(installedPath); err == nil {
		info.Installed = true
		info.InstalledPath = filepath.Dir(installedPath)
	}

	return info, nil
}

// UpdateCache refreshes all cached index files.
func (s *Source) UpdateCache(ctx context.Context) error {
	// Invalidate existing cache
	if err := s.cache.InvalidateAll(); err != nil {
		return fmt.Errorf("invalidating cache: %w", err)
	}

	// Fetch all indexes to repopulate cache
	for _, kind := range []ItemKind{KindSkill, KindPersona, KindProfile} {
		if _, _, err := s.getIndex(ctx, kind); err != nil {
			return fmt.Errorf("fetching %s index: %w", kind.Plural(), err)
		}
	}

	return nil
}
