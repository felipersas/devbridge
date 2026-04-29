package pair

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/felipersas/notifybridge/internal/cfg"
)

func TestGenerateToken(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(token) != 6 {
		t.Errorf("token length = %d, want 6", len(token))
	}
	for _, c := range token {
		if c < '0' || c > '9' {
			t.Errorf("token contains non-digit: %c", c)
		}
	}
}

func TestGenerateTokenUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateToken()
		if err != nil {
			t.Fatal(err)
		}
		if seen[token] {
			t.Errorf("duplicate token: %s", token)
		}
		seen[token] = true
	}
}

func TestReachableIP(t *testing.T) {
	ip, err := reachableIP()
	if err != nil {
		t.Skipf("no reachable IP: %v", err)
	}
	if ip == "" {
		t.Error("expected non-empty IP")
	}
}

func TestEnsureSSHPubKey(t *testing.T) {
	pubKey, err := ensureSSHPubKey()
	if err != nil {
		t.Skipf("SSH key not available: %v", err)
	}
	if pubKey == "" {
		t.Error("expected non-empty public key")
	}
	if len(pubKey) < 50 {
		t.Errorf("pubkey seems too short: %s", pubKey)
	}
}

func TestAddAuthorizedKey_NewKey(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)
	t.Setenv("HOME", tmpDir)

	key := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI testkey@host"
	err := addAuthorizedKey(key)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(sshDir, "authorized_keys"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), key) {
		t.Errorf("authorized_keys should contain %q, got %q", key, string(data))
	}
}

func TestAddAuthorizedKey_DuplicateKey(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)
	t.Setenv("HOME", tmpDir)

	key := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI testkey@host"
	addAuthorizedKey(key)
	addAuthorizedKey(key)

	data, err := os.ReadFile(filepath.Join(sshDir, "authorized_keys"))
	if err != nil {
		t.Fatal(err)
	}
	count := strings.Count(string(data), key)
	if count != 1 {
		t.Errorf("key appears %d times, want 1", count)
	}
}

func TestAddAuthorizedKey_AppendToExisting(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)
	t.Setenv("HOME", tmpDir)

	existingKey := "ssh-ed25519 EXISTINGKEY existing@host"
	os.WriteFile(filepath.Join(sshDir, "authorized_keys"), []byte(existingKey+"\n"), 0600)

	newKey := "ssh-ed25519 NEWKEY new@host"
	addAuthorizedKey(newKey)

	data, err := os.ReadFile(filepath.Join(sshDir, "authorized_keys"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), existingKey) {
		t.Error("should preserve existing key")
	}
	if !strings.Contains(string(data), newKey) {
		t.Error("should contain new key")
	}
}

func TestAddAuthorizedKey_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)
	t.Setenv("HOME", tmpDir)

	// No authorized_keys file exists
	key := "ssh-ed25519 BRANDKEY new@host"
	err := addAuthorizedKey(key)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(sshDir, "authorized_keys"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != key {
		t.Errorf("got %q, want %q", strings.TrimSpace(string(data)), key)
	}
}

func TestWriteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "notifybridge.conf")
	t.Setenv("NOTIFYBRIDGE_CONF", confPath)

	phone := &PhoneInfo{
		IP:         "10.0.0.42",
		SSHPort:    "8022",
		User:       "root",
		DeviceName: "Pixel 8",
	}

	err := writeConfig(phone)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := cfg.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.AndroidIP != "10.0.0.42" {
		t.Errorf("AndroidIP = %q, want %q", loaded.AndroidIP, "10.0.0.42")
	}
	if loaded.SSHUser != "root" {
		t.Errorf("SSHUser = %q, want %q", loaded.SSHUser, "root")
	}
	if loaded.SSHPort != "8022" {
		t.Errorf("SSHPort = %q, want %q", loaded.SSHPort, "8022")
	}
}

func TestOsUserHome(t *testing.T) {
	user, _ := osuserHome()
	if user == "" {
		t.Error("user should not be empty")
	}
}
