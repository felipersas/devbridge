package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.conf")

	content := `# Test projects
DEFAULT=⚡:2196F3:high
Termux=🔧:2196F3:high
licespot-web=🛒:4CAF50:high
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("NOTIFYBRIDGE_PROJECTS", path)

	profiles, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	p, ok := profiles["Termux"]
	if !ok {
		t.Fatal("expected Termux profile")
	}
	if p.Emoji != "🔧" {
		t.Errorf("Termux emoji = %q, want 🔧", p.Emoji)
	}
	if p.LEDColor != "2196F3" {
		t.Errorf("Termux LED = %q, want 2196F3", p.LEDColor)
	}

	p2, ok := profiles["licespot-web"]
	if !ok {
		t.Fatal("expected licespot-web profile")
	}
	if p2.Emoji != "🛒" {
		t.Errorf("licespot-web emoji = %q, want 🛒", p2.Emoji)
	}
}

func TestMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.conf")

	content := `Termux=🔧:2196F3:high
licespot-web=🛒:4CAF50:high
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("NOTIFYBRIDGE_PROJECTS", path)

	p := Match("Termux")
	if p.Emoji != "🔧" {
		t.Errorf("Termux match emoji = %q, want 🔧", p.Emoji)
	}

	p = Match("unknown-project")
	if p.Emoji != DefaultProfile.Emoji {
		t.Errorf("unknown match emoji = %q, want default %q", p.Emoji, DefaultProfile.Emoji)
	}
}

func TestMatchNoFile(t *testing.T) {
	t.Setenv("NOTIFYBRIDGE_PROJECTS", "/nonexistent/path")

	p := Match("anything")
	if p.Emoji != DefaultProfile.Emoji {
		t.Errorf("no file match emoji = %q, want default", p.Emoji)
	}
}
