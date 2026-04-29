package unpair

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/felipersas/devbridge/internal/cfg"
	"github.com/felipersas/devbridge/internal/profile"
)

// Run removes pairing artifacts: config file, projects file, and SSH key pair.
func Run() error {
	return RunWithFS(os.Remove)
}

// RunWithFS accepts a remove function for testing.
func RunWithFS(removeFn func(string) error) error {
	var removed []string
	var errors []string

	// Config file
	confPath, err := cfg.Path()
	if err == nil {
		if err := removeFn(confPath); err == nil {
			removed = append(removed, confPath)
		}
	}

	// Projects file
	projPath, err := profile.ProjectsPath()
	if err == nil {
		if err := removeFn(projPath); err == nil {
			removed = append(removed, projPath)
		}
	}

	// DevBridge SSH key pair
	home, err := os.UserHomeDir()
	if err == nil {
		for _, f := range []string{
			filepath.Join(home, ".ssh", "devbridge_ed25519"),
			filepath.Join(home, ".ssh", "devbridge_ed25519.pub"),
		} {
			if err := removeFn(f); err == nil {
				removed = append(removed, f)
			}
		}
	}

	if len(removed) == 0 {
		return fmt.Errorf("no DevBridge configuration found")
	}

	fmt.Println("  DevBridge Unpair")
	fmt.Println("  ═════════════════")
	fmt.Println()
	for _, f := range removed {
		fmt.Printf("  Removed: %s\n", f)
	}
	if len(errors) > 0 {
		fmt.Println()
		for _, e := range errors {
			fmt.Printf("  Warning: %s\n", e)
		}
	}
	fmt.Println()
	fmt.Println("  Pair again with: devbridge pair")
	return nil
}
