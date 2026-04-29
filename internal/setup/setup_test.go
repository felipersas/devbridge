package setup

import (
	"bufio"
	"path/filepath"
	"strings"
	"testing"

	"github.com/felipersas/devbridge/internal/cfg"
)

func TestAsk_DefaultValue(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("\n"))
	got := ask(r, "Prompt", "default")
	if got != "default" {
		t.Errorf("ask() = %q, want %q", got, "default")
	}
}

func TestAsk_CustomValue(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("custom\n"))
	got := ask(r, "Prompt", "default")
	if got != "custom" {
		t.Errorf("ask() = %q, want %q", got, "custom")
	}
}

func TestAsk_TrimsWhitespace(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("  value  \n"))
	got := ask(r, "Prompt", "default")
	if got != "value" {
		t.Errorf("ask() = %q, want %q", got, "value")
	}
}

func TestAskBool_DefaultTrue(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("\n"))
	got := askBool(r, "Prompt", true)
	if !got {
		t.Error("askBool() = false, want true (default)")
	}
}

func TestAskBool_DefaultFalse(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("\n"))
	got := askBool(r, "Prompt", false)
	if got {
		t.Error("askBool() = true, want false (default)")
	}
}

func TestAskBool_YesInput(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"y\n", true},
		{"yes\n", true},
		{"Y\n", true},
		{"YES\n", true},
		{"n\n", false},
		{"no\n", false},
		{"maybe\n", false},
	}
	for _, tt := range tests {
		r := bufio.NewReader(strings.NewReader(tt.input))
		got := askBool(r, "Prompt", false)
		if got != tt.want {
			t.Errorf("askBool(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRunWithReader_FullFlow(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "devbridge.conf")
	t.Setenv("DEVBRIDGE_CONF", confPath)

	input := "192.168.1.1\ntestuser\n2222\nTest Title\ny\n"
	r := bufio.NewReader(strings.NewReader(input))

	err := RunWithReader(r)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := cfg.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.AndroidIP != "192.168.1.1" {
		t.Errorf("AndroidIP = %q, want %q", loaded.AndroidIP, "192.168.1.1")
	}
	if loaded.SSHUser != "testuser" {
		t.Errorf("SSHUser = %q, want %q", loaded.SSHUser, "testuser")
	}
	if loaded.SSHPort != "2222" {
		t.Errorf("SSHPort = %q, want %q", loaded.SSHPort, "2222")
	}
	if loaded.DefaultTitle != "Test Title" {
		t.Errorf("DefaultTitle = %q, want %q", loaded.DefaultTitle, "Test Title")
	}
	if !loaded.Sound {
		t.Error("Sound = false, want true")
	}
}

func TestRunWithReader_DefaultsUsed(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "devbridge.conf")
	t.Setenv("DEVBRIDGE_CONF", confPath)

	// All empty lines → should use defaults
	input := "\n\n\n\n\n"
	r := bufio.NewReader(strings.NewReader(input))

	err := RunWithReader(r)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := cfg.Load()
	if err != nil {
		t.Fatal(err)
	}

	d := cfg.Default()
	if loaded.AndroidIP != d.AndroidIP {
		t.Errorf("AndroidIP = %q, want default %q", loaded.AndroidIP, d.AndroidIP)
	}
	if loaded.SSHUser != d.SSHUser {
		t.Errorf("SSHUser = %q, want default %q", loaded.SSHUser, d.SSHUser)
	}
}

func TestRunWithReader_SoundNo(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "devbridge.conf")
	t.Setenv("DEVBRIDGE_CONF", confPath)

	input := "\n\n\n\nn\n"
	r := bufio.NewReader(strings.NewReader(input))

	err := RunWithReader(r)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := cfg.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Sound {
		t.Error("Sound = true, want false")
	}
}
