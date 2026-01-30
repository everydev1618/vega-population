package population

import (
	"context"
	"sort"
	"strings"
)

// Search searches across all item types and returns matching results.
func (s *Source) Search(ctx context.Context, query string, opts *SearchOptions) ([]SearchResult, error) {
	var results []SearchResult
	query = strings.ToLower(query)

	kinds := []ItemKind{KindSkill, KindPersona, KindProfile}
	if opts.Kind != "" {
		kinds = []ItemKind{opts.Kind}
	}

	for _, kind := range kinds {
		entries, profiles, err := s.getIndex(ctx, kind)
		if err != nil {
			return nil, err
		}

		if kind == KindProfile {
			for name, entry := range profiles {
				score := calculateProfileScore(query, name, entry, opts.Tags)
				if score > 0 {
					results = append(results, SearchResult{
						Kind:        kind,
						Name:        name,
						Version:     entry.Version,
						Description: entry.Description,
						Tags:        nil, // Profiles don't have tags in the index
						Score:       score,
					})
				}
			}
		} else {
			for name, entry := range entries {
				score := calculateScore(query, name, entry, opts.Tags)
				if score > 0 {
					results = append(results, SearchResult{
						Kind:        kind,
						Name:        name,
						Version:     entry.Version,
						Description: entry.Description,
						Tags:        entry.Tags,
						Score:       score,
					})
				}
			}
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		// Secondary sort by name for stability
		return results[i].Name < results[j].Name
	})

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// calculateScore calculates a relevance score for a search result.
func calculateScore(query, name string, entry IndexEntry, filterTags []string) float64 {
	// Check tag filter first - if tags are specified and don't match, return 0
	if len(filterTags) > 0 {
		hasMatchingTag := false
		for _, filterTag := range filterTags {
			for _, tag := range entry.Tags {
				if strings.EqualFold(tag, filterTag) {
					hasMatchingTag = true
					break
				}
			}
			if hasMatchingTag {
				break
			}
		}
		if !hasMatchingTag {
			return 0
		}
	}

	var score float64
	nameLower := strings.ToLower(name)
	descLower := strings.ToLower(entry.Description)

	// Exact name match
	if nameLower == query {
		score = 1.0
		return score
	}

	// Name contains query
	if strings.Contains(nameLower, query) {
		score = 0.8
	}

	// Tag exact match
	for _, tag := range entry.Tags {
		if strings.EqualFold(tag, query) {
			if score < 0.7 {
				score = 0.7
			}
			break
		}
	}

	// Tag contains query
	for _, tag := range entry.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			if score < 0.6 {
				score = 0.6
			}
			break
		}
	}

	// Description contains query
	if strings.Contains(descLower, query) {
		if score < 0.5 {
			score = 0.5
		}
	}

	return score
}

// calculateProfileScore calculates a relevance score for a profile search result.
func calculateProfileScore(query, name string, entry ProfileIndexEntry, filterTags []string) float64 {
	// Profiles don't have tags in the index, so tag filtering doesn't apply
	if len(filterTags) > 0 {
		return 0
	}

	var score float64
	nameLower := strings.ToLower(name)
	descLower := strings.ToLower(entry.Description)

	// Exact name match
	if nameLower == query {
		score = 1.0
		return score
	}

	// Name contains query
	if strings.Contains(nameLower, query) {
		score = 0.8
	}

	// Description contains query
	if strings.Contains(descLower, query) {
		if score < 0.5 {
			score = 0.5
		}
	}

	// Check if any of the included skills match
	for _, skill := range entry.Skills {
		if strings.Contains(strings.ToLower(skill), query) {
			if score < 0.4 {
				score = 0.4
			}
			break
		}
	}

	// Check if the persona matches
	if strings.Contains(strings.ToLower(entry.Persona), query) {
		if score < 0.4 {
			score = 0.4
		}
	}

	return score
}
