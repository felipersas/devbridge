package health

import (
	"strings"
	"testing"

	"github.com/felipersas/devbridge/internal/cfg"
)

func TestCheck_SSHSuccess_TermuxSuccess(t *testing.T) {
	config := &cfg.Config{
		AndroidIP: "10.0.0.1",
		SSHUser:   "root",
		SSHPort:   "8022",
	}
	stub := func(name string, args ...string) error { return nil }

	r := Check(config, stub)
	if !r.SSHConnected {
		t.Error("SSHConnected = false, want true")
	}
	if !r.TermuxAvailable {
		t.Error("TermuxAvailable = false, want true")
	}
	if r.ErrorMessage != "" {
		t.Errorf("ErrorMessage = %q, want empty", r.ErrorMessage)
	}
}

func TestCheck_SSHFailure(t *testing.T) {
	config := &cfg.Config{
		AndroidIP: "10.0.0.1",
		SSHUser:   "root",
		SSHPort:   "8022",
	}
	stub := func(name string, args ...string) error { return errTest }

	r := Check(config, stub)
	if r.SSHConnected {
		t.Error("SSHConnected = true, want false")
	}
	if r.TermuxAvailable {
		t.Error("TermuxAvailable = true, want false")
	}
	if r.ErrorMessage == "" {
		t.Error("ErrorMessage should not be empty")
	}
}

func TestCheck_SSHSuccess_TermuxFailure(t *testing.T) {
	config := &cfg.Config{
		AndroidIP: "10.0.0.1",
		SSHUser:   "root",
		SSHPort:   "8022",
	}
	callCount := 0
	stub := func(name string, args ...string) error {
		callCount++
		if callCount == 1 {
			return nil // SSH echo OK
		}
		return errTest // termux-notification fails
	}

	r := Check(config, stub)
	if !r.SSHConnected {
		t.Error("SSHConnected = false, want true")
	}
	if r.TermuxAvailable {
		t.Error("TermuxAvailable = true, want false")
	}
}

func TestCheck_UsesCorrectSSHArgs(t *testing.T) {
	config := &cfg.Config{
		AndroidIP: "192.168.1.50",
		SSHUser:   "testuser",
		SSHPort:   "2222",
	}
	stub := func(name string, args ...string) error {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "testuser@192.168.1.50") {
			t.Errorf("args should contain user@host, got: %s", joined)
		}
		if !strings.Contains(joined, "2222") {
			t.Errorf("args should contain port 2222, got: %s", joined)
		}
		return nil
	}

	Check(config, stub)
}

type testError struct{}

func (testError) Error() string { return "test error" }

var errTest error = testError{}
