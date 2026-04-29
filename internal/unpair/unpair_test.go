package unpair

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRun_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("DEVBRIDGE_CONF", filepath.Join(tmpDir, "devbridge.conf"))
	t.Setenv("DEVBRIDGE_PROJECTS", filepath.Join(tmpDir, "devbridge-projects.conf"))
	t.Setenv("HOME", tmpDir)

	err := Run()
	if err == nil {
		t.Fatal("expected error when no config exists")
	}
}

func TestRun_RemovesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "devbridge.conf")
	projPath := filepath.Join(tmpDir, "devbridge-projects.conf")
	t.Setenv("DEVBRIDGE_CONF", confPath)
	t.Setenv("DEVBRIDGE_PROJECTS", projPath)

	os.WriteFile(confPath, []byte("ANDROID_IP=\"10.0.0.1\"\n"), 0644)
	os.WriteFile(projPath, []byte("myproject=:FF0000:high\n"), 0644)

	err := Run()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(confPath); !os.IsNotExist(err) {
		t.Error("config file should be removed")
	}
	if _, err := os.Stat(projPath); !os.IsNotExist(err) {
		t.Error("projects file should be removed")
	}
}

func TestRun_RemovesSSHKey(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)

	keyPath := filepath.Join(sshDir, "devbridge_ed25519")
	pubPath := keyPath + ".pub"
	os.WriteFile(keyPath, []byte("private"), 0600)
	os.WriteFile(pubPath, []byte("public"), 0644)

	t.Setenv("HOME", tmpDir)
	t.Setenv("DEVBRIDGE_CONF", filepath.Join(tmpDir, "devbridge.conf"))
	t.Setenv("DEVBRIDGE_PROJECTS", filepath.Join(tmpDir, "devbridge-projects.conf"))

	// Create a config so something gets removed
	os.WriteFile(filepath.Join(tmpDir, "devbridge.conf"), []byte("ANDROID_IP=\"10.0.0.1\"\n"), 0644)

	err := Run()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		t.Error("SSH private key should be removed")
	}
	if _, err := os.Stat(pubPath); !os.IsNotExist(err) {
		t.Error("SSH public key should be removed")
	}
}

func TestRunWithFS_CustomRemove(t *testing.T) {
	var removed []string
	removeFn := func(path string) error {
		removed = append(removed, path)
		return nil
	}

	tmpDir := t.TempDir()
	t.Setenv("DEVBRIDGE_CONF", filepath.Join(tmpDir, "devbridge.conf"))
	t.Setenv("DEVBRIDGE_PROJECTS", filepath.Join(tmpDir, "devbridge-projects.conf"))
	t.Setenv("HOME", tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "devbridge.conf"), []byte("ANDROID_IP=\"10.0.0.1\"\n"), 0644)

	err := RunWithFS(removeFn)
	if err != nil {
		t.Fatal(err)
	}

	if len(removed) < 1 {
		t.Errorf("expected at least 1 removal, got %d", len(removed))
	}

	// Verify config path was targeted
	found := false
	for _, r := range removed {
		if filepath.Base(r) == "devbridge.conf" {
			found = true
		}
	}
	if !found {
		t.Error("config file should be in removed list")
	}
}

func TestRun_ContinuesOnPartialFailure(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "devbridge.conf")
	t.Setenv("DEVBRIDGE_CONF", confPath)
	t.Setenv("DEVBRIDGE_PROJECTS", filepath.Join(tmpDir, "nonexistent-projects.conf"))

	os.WriteFile(confPath, []byte("ANDROID_IP=\"10.0.0.1\"\n"), 0644)

	// Config exists, projects doesn't — should still succeed
	err := Run()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(confPath); !os.IsNotExist(err) {
		t.Error("config file should be removed even when projects file missing")
	}
}
