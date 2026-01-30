// Package population provides package management for Vega skills, personas, and profiles.
//
// It supports installing, searching, and listing items from the vega-population repository.
// The package can be used as a library by any Go application or through the CLI interface.
//
// Example usage as a library:
//
//	client, err := population.NewClient()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	results, err := client.Search(ctx, "kubernetes", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, r := range results {
//	    fmt.Printf("%s: %s\n", r.Name, r.Description)
//	}
package population

import "strings"

// ItemKind represents the type of population item.
type ItemKind string

const (
	KindSkill   ItemKind = "skill"
	KindPersona ItemKind = "persona"
	KindProfile ItemKind = "profile"
)

// String returns the string representation of the ItemKind.
func (k ItemKind) String() string {
	return string(k)
}

// Plural returns the plural form of the ItemKind (used in paths).
func (k ItemKind) Plural() string {
	switch k {
	case KindSkill:
		return "skills"
	case KindPersona:
		return "personas"
	case KindProfile:
		return "profiles"
	default:
		return string(k) + "s"
	}
}

// SearchResult represents a single search result.
type SearchResult struct {
	Kind        ItemKind
	Name        string
	Version     string
	Description string
	Tags        []string
	Score       float64 // Relevance score 0-1
}

// SearchOptions configures the search behavior.
type SearchOptions struct {
	Kind  ItemKind // Filter by type (empty = all)
	Tags  []string // Filter by tags
	Limit int      // Max results (0 = no limit)
}

// InstallOptions configures the installation behavior.
type InstallOptions struct {
	Force  bool // Overwrite existing installations
	NoDeps bool // Skip profile dependencies (persona and skills)
	DryRun bool // Show what would be installed without actually installing
}

// InstalledItem represents an installed skill, persona, or profile.
type InstalledItem struct {
	Kind    ItemKind
	Name    string
	Version string
	Path    string
}

// ItemInfo contains detailed information about an item.
type ItemInfo struct {
	Kind        ItemKind
	Name        string
	Version     string
	Description string
	Author      string
	Tags        []string
	// For profiles
	Persona string
	Skills  []string
	// For personas
	RecommendedSkills []string
	// Installation status
	Installed     bool
	InstalledPath string
}

// ParseItemName parses an input string and returns the kind and name.
// Names prefixed with @ are personas, + are profiles, and unprefixed are skills.
func ParseItemName(input string) (ItemKind, string) {
	if strings.HasPrefix(input, "@") {
		return KindPersona, strings.TrimPrefix(input, "@")
	}
	if strings.HasPrefix(input, "+") {
		return KindProfile, strings.TrimPrefix(input, "+")
	}
	return KindSkill, input
}

// FormatItemName returns the display name with the appropriate prefix.
func FormatItemName(kind ItemKind, name string) string {
	switch kind {
	case KindPersona:
		return "@" + name
	case KindProfile:
		return "+" + name
	default:
		return name
	}
}
