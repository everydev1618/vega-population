package population

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Install installs an item from the source to the install directory.
func (s *Source) Install(ctx context.Context, kind ItemKind, name string, installDir string, opts *InstallOptions) error {
	// Check if already installed
	destDir := filepath.Join(installDir, kind.Plural(), name)
	destPath := filepath.Join(destDir, "vega.yaml")

	if _, err := os.Stat(destPath); err == nil && !opts.Force {
		return fmt.Errorf("%s %q is already installed (use --force to overwrite)", kind, name)
	}

	if opts.DryRun {
		fmt.Printf("Would install %s %q to %s\n", kind, name, destDir)
	}

	// For profiles, handle dependencies first
	if kind == KindProfile && !opts.NoDeps {
		if err := s.installProfileDeps(ctx, name, installDir, opts); err != nil {
			return err
		}
	}

	// Fetch the manifest
	content, err := s.GetManifestRaw(ctx, kind, name)
	if err != nil {
		return fmt.Errorf("fetching %s %q: %w", kind, name, err)
	}

	if opts.DryRun {
		return nil
	}

	// Create directory and write file
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	return nil
}

// installProfileDeps installs the dependencies of a profile (persona and skills).
func (s *Source) installProfileDeps(ctx context.Context, profileName string, installDir string, opts *InstallOptions) error {
	// Get the profile index to find dependencies
	_, profiles, err := s.getIndex(ctx, KindProfile)
	if err != nil {
		return err
	}

	profile, ok := profiles[profileName]
	if !ok {
		return fmt.Errorf("profile %q not found", profileName)
	}

	// Install persona
	if profile.Persona != "" {
		if opts.DryRun {
			fmt.Printf("Would install persona %q (dependency of profile %q)\n", profile.Persona, profileName)
		} else {
			fmt.Printf("Installing persona %q...\n", profile.Persona)
		}

		depOpts := &InstallOptions{
			Force:  opts.Force,
			NoDeps: true, // Don't recurse for personas
			DryRun: opts.DryRun,
		}

		if err := s.Install(ctx, KindPersona, profile.Persona, installDir, depOpts); err != nil {
			// Don't fail on "already installed" errors for dependencies
			if !opts.Force && isAlreadyInstalledError(err) {
				if !opts.DryRun {
					fmt.Printf("  Persona %q already installed\n", profile.Persona)
				}
			} else {
				return fmt.Errorf("installing persona %q: %w", profile.Persona, err)
			}
		}
	}

	// Install skills
	for _, skillName := range profile.Skills {
		if opts.DryRun {
			fmt.Printf("Would install skill %q (dependency of profile %q)\n", skillName, profileName)
		} else {
			fmt.Printf("Installing skill %q...\n", skillName)
		}

		depOpts := &InstallOptions{
			Force:  opts.Force,
			NoDeps: true,
			DryRun: opts.DryRun,
		}

		if err := s.Install(ctx, KindSkill, skillName, installDir, depOpts); err != nil {
			if !opts.Force && isAlreadyInstalledError(err) {
				if !opts.DryRun {
					fmt.Printf("  Skill %q already installed\n", skillName)
				}
			} else {
				return fmt.Errorf("installing skill %q: %w", skillName, err)
			}
		}
	}

	return nil
}

// isAlreadyInstalledError checks if the error is an "already installed" error.
func isAlreadyInstalledError(err error) bool {
	if err == nil {
		return false
	}
	return containsString(err.Error(), "already installed")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
