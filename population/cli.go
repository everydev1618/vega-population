package population

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RunCLI is the entry point for the CLI interface.
func RunCLI(args []string) error {
	if len(args) == 0 {
		return printUsage()
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "search":
		return runSearch(cmdArgs)
	case "install":
		return runInstall(cmdArgs)
	case "list", "ls":
		return runList(cmdArgs)
	case "info":
		return runInfo(cmdArgs)
	case "export":
		return runExport(cmdArgs)
	case "sync":
		return runSync(cmdArgs)
	case "reindex":
		return runReindex(cmdArgs)
	case "update":
		return runUpdate(cmdArgs)
	case "help", "-h", "--help":
		return printUsage()
	default:
		return fmt.Errorf("unknown command: %s\nRun 'vega population help' for usage", cmd)
	}
}

func printUsage() error {
	fmt.Println(`Usage: vega population <command> [options]

Commands:
  search <query>     Search for skills, personas, and profiles
  install <name>     Install a skill, persona (@name), or profile (+name)
  list               List installed items
  info <name>        Show detailed information about an item
  export <name>      Export a persona as YAML for tron.vega.yaml
  sync               Sync personas as Claude Code skills (~/.claude/skills/)
  reindex            Regenerate index.yaml files from directories
  update             Update the local cache

Examples:
  vega population search kubernetes
  vega population install kubernetes-ops
  vega population install @incident-commander
  vega population install +platform-engineer
  vega population export @cmo
  vega population sync
  vega population sync @ceo @cto
  vega population list`)
	return nil
}

func runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	kindFlag := fs.String("kind", "", "Filter by kind (skill, persona, profile)")
	tagsFlag := fs.String("tags", "", "Filter by tags (comma-separated)")
	limitFlag := fs.Int("limit", 0, "Maximum number of results")
	sourceFlag := fs.String("source", "", "Custom source URL or path")
	noCacheFlag := fs.Bool("no-cache", false, "Disable caching")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("search requires a query argument")
	}

	query := strings.Join(fs.Args(), " ")

	var opts []Option
	if *sourceFlag != "" {
		opts = append(opts, WithSource(*sourceFlag))
	}
	if *noCacheFlag {
		opts = append(opts, WithNoCache())
	}

	client, err := NewClient(opts...)
	if err != nil {
		return err
	}

	searchOpts := &SearchOptions{
		Limit: *limitFlag,
	}

	if *kindFlag != "" {
		searchOpts.Kind = ItemKind(*kindFlag)
	}

	if *tagsFlag != "" {
		searchOpts.Tags = strings.Split(*tagsFlag, ",")
		for i, t := range searchOpts.Tags {
			searchOpts.Tags[i] = strings.TrimSpace(t)
		}
	}

	results, err := client.Search(context.Background(), query, searchOpts)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Printf("No results found for %q\n", query)
		return nil
	}

	fmt.Printf("Found %d result(s) for %q:\n\n", len(results), query)

	for _, r := range results {
		name := FormatItemName(r.Kind, r.Name)
		fmt.Printf("  %-30s  %s\n", name, r.Description)
		if len(r.Tags) > 0 {
			fmt.Printf("  %-30s  tags: %s\n", "", strings.Join(r.Tags, ", "))
		}
		fmt.Println()
	}

	return nil
}

func runInstall(args []string) error {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	forceFlag := fs.Bool("force", false, "Overwrite existing installation")
	noDepsFlag := fs.Bool("no-deps", false, "Skip profile dependencies")
	dryRunFlag := fs.Bool("dry-run", false, "Show what would be installed")
	sourceFlag := fs.String("source", "", "Custom source URL or path")
	installDirFlag := fs.String("install-dir", "", "Custom installation directory")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("install requires a name argument")
	}

	var opts []Option
	if *sourceFlag != "" {
		opts = append(opts, WithSource(*sourceFlag))
	}
	if *installDirFlag != "" {
		opts = append(opts, WithInstallDir(*installDirFlag))
	}

	client, err := NewClient(opts...)
	if err != nil {
		return err
	}

	installOpts := &InstallOptions{
		Force:  *forceFlag,
		NoDeps: *noDepsFlag,
		DryRun: *dryRunFlag,
	}

	for _, name := range fs.Args() {
		kind, itemName := ParseItemName(name)

		if !*dryRunFlag {
			fmt.Printf("Installing %s %q...\n", kind, itemName)
		}

		if err := client.Install(context.Background(), name, installOpts); err != nil {
			return err
		}

		if !*dryRunFlag {
			fmt.Printf("Successfully installed %s to %s/%s/%s\n", FormatItemName(kind, itemName), client.InstallDir(), kind.Plural(), itemName)
		}
	}

	return nil
}

func runList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	kindFlag := fs.String("kind", "", "Filter by kind (skill, persona, profile)")
	installDirFlag := fs.String("install-dir", "", "Custom installation directory")

	if err := fs.Parse(args); err != nil {
		return err
	}

	var opts []Option
	if *installDirFlag != "" {
		opts = append(opts, WithInstallDir(*installDirFlag))
	}

	client, err := NewClient(opts...)
	if err != nil {
		return err
	}

	var kind ItemKind
	if *kindFlag != "" {
		kind = ItemKind(*kindFlag)
	}

	items, err := client.List(kind)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("No items installed")
		return nil
	}

	// Group by kind
	byKind := make(map[ItemKind][]InstalledItem)
	for _, item := range items {
		byKind[item.Kind] = append(byKind[item.Kind], item)
	}

	for _, k := range []ItemKind{KindSkill, KindPersona, KindProfile} {
		items, ok := byKind[k]
		if !ok {
			continue
		}

		fmt.Printf("%s:\n", titleCase(k.Plural()))
		for _, item := range items {
			name := FormatItemName(item.Kind, item.Name)
			fmt.Printf("  %-30s  v%s\n", name, item.Version)
		}
		fmt.Println()
	}

	return nil
}

func runInfo(args []string) error {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	sourceFlag := fs.String("source", "", "Custom source URL or path")
	installDirFlag := fs.String("install-dir", "", "Custom installation directory")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("info requires a name argument")
	}

	var opts []Option
	if *sourceFlag != "" {
		opts = append(opts, WithSource(*sourceFlag))
	}
	if *installDirFlag != "" {
		opts = append(opts, WithInstallDir(*installDirFlag))
	}

	client, err := NewClient(opts...)
	if err != nil {
		return err
	}

	name := fs.Arg(0)
	info, err := client.Info(context.Background(), name)
	if err != nil {
		return err
	}

	fmt.Printf("Name:        %s\n", FormatItemName(info.Kind, info.Name))
	fmt.Printf("Kind:        %s\n", info.Kind)
	fmt.Printf("Version:     %s\n", info.Version)
	fmt.Printf("Description: %s\n", info.Description)
	fmt.Printf("Author:      %s\n", info.Author)

	if len(info.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(info.Tags, ", "))
	}

	if info.Persona != "" {
		fmt.Printf("Persona:     @%s\n", info.Persona)
	}

	if len(info.Skills) > 0 {
		fmt.Printf("Skills:      %s\n", strings.Join(info.Skills, ", "))
	}

	if len(info.RecommendedSkills) > 0 {
		fmt.Printf("Recommended: %s\n", strings.Join(info.RecommendedSkills, ", "))
	}

	fmt.Println()
	if info.Installed {
		fmt.Printf("Status:      Installed at %s\n", info.InstalledPath)
	} else {
		fmt.Printf("Status:      Not installed\n")
	}

	return nil
}

func runExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	sourceFlag := fs.String("source", "", "Custom source URL or path")
	nameFlag := fs.String("name", "", "Agent name to use (default: extracted from persona or capitalized ID)")
	modelFlag := fs.String("model", "claude-sonnet-4-20250514", "Model to use")
	tempFlag := fs.Float64("temperature", 0.7, "Temperature setting")
	budgetFlag := fs.String("budget", "$3.00", "Budget limit")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("export requires a persona name (e.g., @cmo)")
	}

	name := fs.Arg(0)
	kind, itemName := ParseItemName(name)

	if kind != KindPersona {
		return fmt.Errorf("export only works with personas (use @name format)")
	}

	var opts []Option
	if *sourceFlag != "" {
		opts = append(opts, WithSource(*sourceFlag))
	}

	client, err := NewClient(opts...)
	if err != nil {
		return err
	}

	source := NewSource(client.source, client.cache)

	// Fetch the manifest
	manifest, err := source.GetManifest(context.Background(), kind, itemName)
	if err != nil {
		return fmt.Errorf("fetching persona: %w", err)
	}

	// Determine agent name
	agentName := *nameFlag
	if agentName == "" {
		// Try to extract name from "You are X" in system prompt
		agentName = extractAgentName(manifest.SystemPrompt)
		if agentName == "" {
			agentName = titleCase(itemName)
		}
	}

	// Output in tron.vega.yaml format
	fmt.Printf("  %s:\n", agentName)
	fmt.Printf("    model: %s\n", *modelFlag)
	fmt.Printf("    temperature: %v\n", *tempFlag)
	fmt.Printf("    budget: \"%s\"\n", *budgetFlag)
	fmt.Printf("    system: |\n")

	// Indent the system prompt
	lines := strings.Split(manifest.SystemPrompt, "\n")
	for _, line := range lines {
		fmt.Printf("      %s\n", line)
	}

	fmt.Printf("    tools:\n")
	fmt.Printf("      - read_file\n")
	fmt.Printf("      - write_file\n")
	fmt.Printf("      - web_search\n")
	fmt.Printf("    supervision:\n")
	fmt.Printf("      strategy: restart\n")
	fmt.Printf("      max_restarts: 2\n")

	return nil
}

func runSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	sourceFlag := fs.String("source", "", "Custom source URL or path")
	targetFlag := fs.String("target", "", "Target directory (default: ~/.claude/skills)")
	projectFlag := fs.Bool("project", false, "Sync to .claude/skills/ in current directory")
	prefixFlag := fs.String("prefix", "vega-", "Prefix for skill directory names")
	pruneFlag := fs.Bool("prune", false, "Remove skills no longer in population source")
	forceFlag := fs.Bool("force", false, "Overwrite even if unchanged")
	dryRunFlag := fs.Bool("dry-run", false, "Show what would be done without making changes")
	noCacheFlag := fs.Bool("no-cache", false, "Disable caching")

	if err := fs.Parse(args); err != nil {
		return err
	}

	var opts []Option
	if *sourceFlag != "" {
		opts = append(opts, WithSource(*sourceFlag))
	}
	if *noCacheFlag {
		opts = append(opts, WithNoCache())
	}

	client, err := NewClient(opts...)
	if err != nil {
		return err
	}

	syncOpts := &SyncOptions{
		Prefix: *prefixFlag,
		Prune:  *pruneFlag,
		Force:  *forceFlag,
		DryRun: *dryRunFlag,
	}

	// Determine target directory
	if *targetFlag != "" {
		syncOpts.TargetDir = *targetFlag
	} else if *projectFlag {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine current directory: %w", err)
		}
		syncOpts.TargetDir = filepath.Join(cwd, ".claude", "skills")
	}

	// Collect specific persona names from positional args
	for _, arg := range fs.Args() {
		syncOpts.Personas = append(syncOpts.Personas, arg)
	}

	if !*dryRunFlag {
		fmt.Printf("Syncing personas to Claude Code skills...\n")
		if syncOpts.TargetDir != "" {
			fmt.Printf("Target: %s\n", syncOpts.TargetDir)
		} else {
			fmt.Printf("Target: ~/.claude/skills/\n")
		}
		fmt.Println()
	}

	result, err := client.Sync(context.Background(), syncOpts)
	if err != nil {
		return err
	}

	// Print results
	if len(result.Created) > 0 {
		fmt.Printf("Created %d skill(s):\n", len(result.Created))
		for _, name := range result.Created {
			fmt.Printf("  + %s%s/SKILL.md\n", syncOpts.Prefix, name)
		}
	}

	if len(result.Updated) > 0 {
		fmt.Printf("Updated %d skill(s):\n", len(result.Updated))
		for _, name := range result.Updated {
			fmt.Printf("  ~ %s%s/SKILL.md\n", syncOpts.Prefix, name)
		}
	}

	if len(result.Pruned) > 0 {
		fmt.Printf("Pruned %d skill(s):\n", len(result.Pruned))
		for _, name := range result.Pruned {
			fmt.Printf("  - %s%s/\n", syncOpts.Prefix, name)
		}
	}

	if len(result.Skipped) > 0 && !*dryRunFlag {
		fmt.Printf("Skipped %d unchanged skill(s)\n", len(result.Skipped))
	}

	total := len(result.Created) + len(result.Updated) + len(result.Skipped)
	if !*dryRunFlag && total > 0 {
		fmt.Printf("\nDone. %d persona(s) available as Claude Code skills.\n", total)
	}

	return nil
}

func runReindex(args []string) error {
	fs := flag.NewFlagSet("reindex", flag.ExitOnError)
	rootFlag := fs.String("root", "", "Root directory of vega-population repo (default: cwd)")
	checkFlag := fs.Bool("check", false, "Check if indexes are up to date (for CI, exits non-zero if stale)")
	dryRunFlag := fs.Bool("dry-run", false, "Show what would change without writing")

	if err := fs.Parse(args); err != nil {
		return err
	}

	opts := &ReindexOptions{
		RootDir: *rootFlag,
		Check:   *checkFlag,
		DryRun:  *dryRunFlag,
	}

	result, err := Reindex(opts)
	if err != nil {
		return err
	}

	if result.UpToDate {
		fmt.Println("All indexes are up to date.")
		return nil
	}

	// Print changes
	printReindexChanges("personas", result.PersonasAdded, result.PersonasRemoved)
	printReindexChanges("skills", result.SkillsAdded, result.SkillsRemoved)
	printReindexChanges("profiles", result.ProfilesAdded, result.ProfilesRemoved)

	if *checkFlag {
		return fmt.Errorf("indexes are out of date — run: vega population reindex")
	}

	if *dryRunFlag {
		fmt.Println("\nDry run — no files written.")
	} else {
		fmt.Println("\nIndexes updated.")
	}

	return nil
}

func printReindexChanges(kind string, added, removed []string) {
	for _, name := range added {
		fmt.Printf("  + %s/%s (new)\n", kind, name)
	}
	for _, name := range removed {
		fmt.Printf("  - %s/%s (removed)\n", kind, name)
	}
}

func runUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	sourceFlag := fs.String("source", "", "Custom source URL or path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	var opts []Option
	if *sourceFlag != "" {
		opts = append(opts, WithSource(*sourceFlag))
	}

	client, err := NewClient(opts...)
	if err != nil {
		return err
	}

	fmt.Println("Updating cache...")
	if err := client.UpdateCache(context.Background()); err != nil {
		return err
	}

	fmt.Println("Cache updated successfully")
	return nil
}

// titleCase returns the string with the first letter capitalized.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// extractAgentName tries to extract a name from "You are X" in the system prompt.
func extractAgentName(systemPrompt string) string {
	// Look for patterns like "You are Maya" or "You are Maya,"
	lines := strings.Split(systemPrompt, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "You are ") {
			// Extract the name
			rest := strings.TrimPrefix(line, "You are ")
			// Take the first word (the name)
			parts := strings.FieldsFunc(rest, func(r rune) bool {
				return r == ' ' || r == ',' || r == '.' || r == '-' || r == ':'
			})
			if len(parts) > 0 {
				name := parts[0]
				// Skip articles
				if name == "a" || name == "an" || name == "the" {
					if len(parts) > 1 {
						return ""
					}
				}
				return name
			}
		}
	}
	return ""
}
