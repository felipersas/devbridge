package profile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Profile maps a project directory to notification settings.
type Profile struct {
	Emoji    string
	LEDColor string
	Priority string
}

// DefaultProfile is the fallback when no project matches.
var DefaultProfile = Profile{
	Emoji:    "⚡",
	LEDColor: "2196F3",
	Priority: "high",
}

// ProjectsPath returns the project profiles config path.
func ProjectsPath() (string, error) {
	if env := os.Getenv("NOTIFYBRIDGE_PROJECTS"); env != "" {
		return env, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	return filepath.Join(home, ".notifybridge-projects.conf"), nil
}

// Load reads all project profiles from disk.
func Load() (map[string]Profile, error) {
	path, err := ProjectsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read projects %s: %w", path, err)
	}

	profiles := make(map[string]Profile)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		parts := strings.SplitN(value, ":", 3)
		if len(parts) != 3 {
			continue
		}

		profiles[key] = Profile{
			Emoji:    parts[0],
			LEDColor: parts[1],
			Priority: parts[2],
		}
	}

	return profiles, scanner.Err()
}

// Match finds the profile for a directory name, falling back to default.
func Match(dirName string) Profile {
	profiles, err := Load()
	if err != nil {
		return DefaultProfile
	}
	if p, ok := profiles[dirName]; ok {
		return p
	}
	return DefaultProfile
}
