package cfg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseKV(t *testing.T) {
	tests := []struct {
		line    string
		wantKey string
		wantVal string
		wantOK  bool
	}{
		{`ANDROID_IP="100.115.83.120"`, "ANDROID_IP", "100.115.83.120", true},
		{"SSH_USER=root", "SSH_USER", "root", true},
		{"SOUND=true", "SOUND", "true", true},
		{"# comment", "", "", false},
		{"", "", "", false},
		{"NOEQUALSSIGN", "", "", false},
	}

	for _, tt := range tests {
		key, val, ok := parseKV(tt.line)
		if ok != tt.wantOK || key != tt.wantKey || val != tt.wantVal {
			t.Errorf("parseKV(%q) = (%q, %q, %v), want (%q, %q, %v)",
				tt.line, key, val, ok, tt.wantKey, tt.wantVal, tt.wantOK)
		}
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.conf")

	content := `# Test config
ANDROID_IP="192.168.1.1"
SSH_USER=testuser
SSH_PORT=2222
DEFAULT_TITLE="Test Title"
SOUND=false
PRIORITY=low
MAX_RETRIES=5
RETRY_DELAY=10
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DEVBRIDGE_CONF", path)

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if c.AndroidIP != "192.168.1.1" {
		t.Errorf("AndroidIP = %q, want %q", c.AndroidIP, "192.168.1.1")
	}
	if c.SSHUser != "testuser" {
		t.Errorf("SSHUser = %q, want %q", c.SSHUser, "testuser")
	}
	if c.SSHPort != "2222" {
		t.Errorf("SSHPort = %q, want %q", c.SSHPort, "2222")
	}
	if c.DefaultTitle != "Test Title" {
		t.Errorf("DefaultTitle = %q, want %q", c.DefaultTitle, "Test Title")
	}
	if c.Sound {
		t.Error("Sound = true, want false")
	}
	if c.Priority != "low" {
		t.Errorf("Priority = %q, want %q", c.Priority, "low")
	}
	if c.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", c.MaxRetries)
	}
	if c.RetryDelay != 10 {
		t.Errorf("RetryDelay = %d, want 10", c.RetryDelay)
	}
}

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "minimal.conf")

	if err := os.WriteFile(path, []byte(`ANDROID_IP="10.0.0.1"`), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DEVBRIDGE_CONF", path)

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if c.AndroidIP != "10.0.0.1" {
		t.Errorf("AndroidIP = %q, want %q", c.AndroidIP, "10.0.0.1")
	}
	if c.SSHUser != "root" {
		t.Errorf("SSHUser = %q, want default 'root'", c.SSHUser)
	}
	if c.MaxRetries != 2 {
		t.Errorf("MaxRetries = %d, want default 2", c.MaxRetries)
	}
}

func TestWriteAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.conf")

	original := &Config{
		AndroidIP:    "1.2.3.4",
		SSHUser:      "admin",
		SSHPort:      "2222",
		DefaultTitle: "Test",
		Sound:        true,
		Priority:     "max",
		MaxRetries:   3,
		RetryDelay:   5,
	}

	if err := Write(path, original); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DEVBRIDGE_CONF", path)

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.AndroidIP != original.AndroidIP {
		t.Errorf("AndroidIP = %q, want %q", loaded.AndroidIP, original.AndroidIP)
	}
	if loaded.SSHUser != original.SSHUser {
		t.Errorf("SSHUser = %q, want %q", loaded.SSHUser, original.SSHUser)
	}
	if loaded.MaxRetries != original.MaxRetries {
		t.Errorf("MaxRetries = %d, want %d", loaded.MaxRetries, original.MaxRetries)
	}
	if loaded.RetryDelay != original.RetryDelay {
		t.Errorf("RetryDelay = %d, want %d", loaded.RetryDelay, original.RetryDelay)
	}
}
