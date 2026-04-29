package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/felipersas/devbridge/internal/cfg"
	"github.com/felipersas/devbridge/internal/notify"
)

type stubNotifier struct {
	last  notify.Notification
	calls int
}

func (s *stubNotifier) Send(n notify.Notification) error {
	s.calls++
	s.last = n
	return nil
}

func (s *stubNotifier) SendBackground(n notify.Notification) {
	s.calls++
	s.last = n
}

func writeTestConfig(t *testing.T, tmpDir string) {
	t.Helper()
	confContent := `ANDROID_IP="10.0.0.1"
SSH_USER="testuser"
SSH_PORT="2222"
DEFAULT_TITLE="Test"
SOUND=true
PRIORITY="high"
MAX_RETRIES=1
RETRY_DELAY=0
`
	confPath := filepath.Join(tmpDir, "devbridge.conf")
	if err := os.WriteFile(confPath, []byte(confContent), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DEVBRIDGE_CONF", confPath)
}

func writeTestProjects(t *testing.T, tmpDir string) {
	t.Helper()
	projectsContent := "DEFAULT=⚡:2196F3:high\nmyproject=🛒:4CAF50:high\n"
	projectsPath := filepath.Join(tmpDir, "projects.conf")
	if err := os.WriteFile(projectsPath, []byte(projectsContent), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DEVBRIDGE_PROJECTS", projectsPath)
}

func TestRunWith_ValidInput(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestConfig(t, tmpDir)
	writeTestProjects(t, tmpDir)

	config, err := cfg.Load()
	if err != nil {
		t.Fatal(err)
	}

	input := `{"cwd": "/home/user/myproject", "last_assistant_message": "Hello from Claude"}`
	stub := &stubNotifier{}

	err = RunWith(strings.NewReader(input), config, stub)
	if err != nil {
		t.Fatal(err)
	}
	if stub.calls != 1 {
		t.Errorf("calls = %d, want 1", stub.calls)
	}
	if stub.last.Title != "🛒 Claude Code — myproject" {
		t.Errorf("Title = %q, want %q", stub.last.Title, "🛒 Claude Code — myproject")
	}
	if stub.last.Message != "Hello from Claude" {
		t.Errorf("Message = %q, want %q", stub.last.Message, "Hello from Claude")
	}
	if stub.last.Group != "claude-myproject" {
		t.Errorf("Group = %q, want %q", stub.last.Group, "claude-myproject")
	}
}

func TestRunWith_EmptyMessage(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestConfig(t, tmpDir)
	writeTestProjects(t, tmpDir)

	config, _ := cfg.Load()
	input := `{"cwd": "/home/user/test", "last_assistant_message": ""}`
	stub := &stubNotifier{}

	err := RunWith(strings.NewReader(input), config, stub)
	if err != nil {
		t.Fatal(err)
	}
	if stub.last.Message != "Finished — needs your input" {
		t.Errorf("Message = %q, want fallback", stub.last.Message)
	}
}

func TestRunWith_MessageTruncation(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestConfig(t, tmpDir)
	writeTestProjects(t, tmpDir)

	config, _ := cfg.Load()
	longMsg := strings.Repeat("a", 300)
	input := `{"cwd": "/home/user/test", "last_assistant_message": "` + longMsg + `"}`
	stub := &stubNotifier{}

	err := RunWith(strings.NewReader(input), config, stub)
	if err != nil {
		t.Fatal(err)
	}
	if len(stub.last.Message) != 200 {
		t.Errorf("Message length = %d, want 200", len(stub.last.Message))
	}
	if !strings.HasSuffix(stub.last.Message, "...") {
		t.Errorf("Message should end with ..., got %q", stub.last.Message)
	}
}

func TestRunWith_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestConfig(t, tmpDir)

	config, _ := cfg.Load()
	stub := &stubNotifier{}

	err := RunWith(strings.NewReader("not json"), config, stub)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse hook input") {
		t.Errorf("error = %q, want parse hook input", err.Error())
	}
}

func TestRunWith_DefaultProfile(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestConfig(t, tmpDir)
	writeTestProjects(t, tmpDir)

	config, _ := cfg.Load()
	input := `{"cwd": "/home/user/unknown-dir", "last_assistant_message": "test"}`
	stub := &stubNotifier{}

	err := RunWith(strings.NewReader(input), config, stub)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stub.last.Title, "⚡") {
		t.Errorf("Title = %q, should use default emoji ⚡", stub.last.Title)
	}
}

func TestRunWith_EmptyCWD(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestConfig(t, tmpDir)
	writeTestProjects(t, tmpDir)

	config, _ := cfg.Load()
	input := `{"cwd": "", "last_assistant_message": "test"}`
	stub := &stubNotifier{}

	err := RunWith(strings.NewReader(input), config, stub)
	if err != nil {
		t.Fatal(err)
	}
	// Should use os.Getwd() as fallback
	wd, _ := os.Getwd()
	expectedDir := filepath.Base(wd)
	if !strings.Contains(stub.last.Title, expectedDir) {
		t.Errorf("Title = %q, should contain %q", stub.last.Title, expectedDir)
	}
}
