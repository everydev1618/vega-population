package population

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SyncOptions configures the sync behavior.
type SyncOptions struct {
	// Target directory for Claude skills. Defaults to ~/.claude/skills.
	TargetDir string
	// Prefix for skill directory names. Defaults to "vega-".
	Prefix string
	// Prune removes skills that are no longer in the population source.
	Prune bool
	// DryRun shows what would be done without making changes.
	DryRun bool
	// Force overwrites existing skills even if unchanged.
	Force bool
	// Personas is an optional list of specific persona names to sync.
	// If empty, all personas are synced.
	Personas []string
}

// SyncResult contains the results of a sync operation.
type SyncResult struct {
	Created []string
	Updated []string
	Pruned  []string
	Skipped []string
}

// syncManifest tracks which personas were synced by vega.
type syncManifest struct {
	Source   string   `yaml:"source"`
	Personas []string `yaml:"personas"`
}

const syncManifestFile = ".vega-sync.yaml"

// Sync exports personas from the population source as Claude Code skills.
func (c *Client) Sync(ctx context.Context, opts *SyncOptions) (*SyncResult, error) {
	if opts == nil {
		opts = &SyncOptions{}
	}

	if opts.Prefix == "" {
		opts.Prefix = "vega-"
	}

	if opts.TargetDir == "" {
		// Respect CLAUDE_CONFIG_DIR if set (multi-profile support)
		claudeDir := os.Getenv("CLAUDE_CONFIG_DIR")
		if claudeDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("could not determine home directory: %w", err)
			}
			claudeDir = filepath.Join(home, ".claude")
		}
		opts.TargetDir = filepath.Join(claudeDir, "skills")
	}

	source := NewSource(c.source, c.cache)

	// Get personas index
	entries, _, err := source.getIndex(ctx, KindPersona)
	if err != nil {
		return nil, fmt.Errorf("fetching personas index: %w", err)
	}

	// Filter to specific personas if requested
	if len(opts.Personas) > 0 {
		filtered := make(map[string]IndexEntry)
		for _, name := range opts.Personas {
			// Strip @ prefix if present
			name = strings.TrimPrefix(name, "@")
			entry, ok := entries[name]
			if !ok {
				return nil, fmt.Errorf("persona %q not found", name)
			}
			filtered[name] = entry
		}
		entries = filtered
	}

	result := &SyncResult{}
	var syncedNames []string

	for name := range entries {
		skillDir := filepath.Join(opts.TargetDir, opts.Prefix+name)
		skillFile := filepath.Join(skillDir, "SKILL.md")

		// Fetch the full manifest to get the system prompt
		manifest, err := source.GetManifest(ctx, KindPersona, name)
		if err != nil {
			return nil, fmt.Errorf("fetching persona %q: %w", name, err)
		}

		// Generate SKILL.md content
		content := generateSkillContent(manifest, opts.Prefix)

		// Check if skill already exists and is unchanged
		existing, err := os.ReadFile(skillFile)
		if err == nil && string(existing) == content && !opts.Force {
			result.Skipped = append(result.Skipped, name)
			syncedNames = append(syncedNames, name)
			continue
		}

		isUpdate := err == nil

		if opts.DryRun {
			if isUpdate {
				fmt.Printf("Would update %s\n", skillDir)
				result.Updated = append(result.Updated, name)
			} else {
				fmt.Printf("Would create %s\n", skillDir)
				result.Created = append(result.Created, name)
			}
			syncedNames = append(syncedNames, name)
			continue
		}

		// Create directory and write SKILL.md
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return nil, fmt.Errorf("creating skill directory %s: %w", skillDir, err)
		}

		if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("writing skill file %s: %w", skillFile, err)
		}

		if isUpdate {
			result.Updated = append(result.Updated, name)
		} else {
			result.Created = append(result.Created, name)
		}
		syncedNames = append(syncedNames, name)
	}

	// Handle pruning
	if opts.Prune {
		pruned, err := pruneStaleSkills(opts.TargetDir, opts.Prefix, syncedNames, opts.DryRun)
		if err != nil {
			return nil, fmt.Errorf("pruning stale skills: %w", err)
		}
		result.Pruned = pruned
	}

	// Write sync manifest (unless dry run)
	if !opts.DryRun {
		if err := writeSyncManifest(opts.TargetDir, c.source, syncedNames); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write sync manifest: %v\n", err)
		}
	}

	return result, nil
}

// generateSkillContent creates the SKILL.md content from a persona manifest.
func generateSkillContent(m *Manifest, prefix string) string {
	var b strings.Builder

	// Frontmatter
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("name: %s%s\n", prefix, m.Name))

	// Build description from persona description + tags hint
	desc := m.Description
	if len(desc) > 0 {
		// Capitalize first letter
		desc = strings.ToUpper(desc[:1]) + desc[1:]
	}
	b.WriteString(fmt.Sprintf("description: \"%s\"\n", escapeYAMLString(desc)))
	b.WriteString("---\n\n")

	// System prompt as the skill body
	b.WriteString(strings.TrimRight(m.SystemPrompt, "\n"))
	b.WriteString("\n")

	// Append system_prompt_append if present
	if m.SystemPromptAppend != "" {
		b.WriteString("\n")
		b.WriteString(strings.TrimRight(m.SystemPromptAppend, "\n"))
		b.WriteString("\n")
	}

	return b.String()
}

// escapeYAMLString escapes quotes in a YAML string value.
func escapeYAMLString(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

// pruneStaleSkills removes vega-prefixed skill directories that are no longer in the source.
func pruneStaleSkills(targetDir, prefix string, currentNames []string, dryRun bool) ([]string, error) {
	// Build set of current names
	current := make(map[string]bool, len(currentNames))
	for _, name := range currentNames {
		current[prefix+name] = true
	}

	// Read existing vega-prefixed directories
	entries, err := os.ReadDir(targetDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var pruned []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if current[name] {
			continue
		}

		// This is a stale vega skill
		dirPath := filepath.Join(targetDir, name)
		if dryRun {
			fmt.Printf("Would prune %s\n", dirPath)
		} else {
			if err := os.RemoveAll(dirPath); err != nil {
				return nil, fmt.Errorf("removing %s: %w", dirPath, err)
			}
		}
		pruned = append(pruned, strings.TrimPrefix(name, prefix))
	}

	return pruned, nil
}

// writeSyncManifest writes a manifest tracking what was synced.
func writeSyncManifest(targetDir, source string, personas []string) error {
	manifest := syncManifest{
		Source:   source,
		Personas: personas,
	}

	data, err := yaml.Marshal(&manifest)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(targetDir, syncManifestFile), data, 0644)
}
