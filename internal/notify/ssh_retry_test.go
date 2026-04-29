package notify

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/felipersas/devbridge/internal/cfg"
)

func TestNewSSHNotifier(t *testing.T) {
	c := &cfg.Config{AndroidIP: "1.2.3.4", SSHUser: "root", SSHPort: "8022"}
	n := NewSSHNotifier(c)
	if n == nil {
		t.Fatal("notifier is nil")
	}
	if n.config != c {
		t.Error("config not set")
	}
	if n.runCmd == nil {
		t.Error("runCmd not initialized")
	}
}

func TestSend_Success(t *testing.T) {
	c := &cfg.Config{
		AndroidIP:  "1.2.3.4",
		SSHUser:    "root",
		SSHPort:    "8022",
		MaxRetries: 1,
		RetryDelay: 0,
	}
	n := NewSSHNotifier(c)
	n.runCmd = func(name string, args ...string) error { return nil }

	err := n.Send(Notification{Title: "T", Message: "M"})
	if err != nil {
		t.Errorf("Send() error = %v", err)
	}
}

func TestSend_RetrySuccess(t *testing.T) {
	var calls atomic.Int32
	c := &cfg.Config{
		AndroidIP:  "1.2.3.4",
		SSHUser:    "root",
		SSHPort:    "8022",
		MaxRetries: 3,
		RetryDelay: 0,
	}
	n := NewSSHNotifier(c)
	n.runCmd = func(name string, args ...string) error {
		if calls.Add(1) < 3 {
			return errFake
		}
		return nil
	}

	err := n.Send(Notification{Title: "T", Message: "M"})
	if err != nil {
		t.Errorf("Send() error = %v", err)
	}
	if calls.Load() != 3 {
		t.Errorf("calls = %d, want 3", calls.Load())
	}
}

func TestSend_AllAttemptsFail(t *testing.T) {
	var calls atomic.Int32
	c := &cfg.Config{
		AndroidIP:  "1.2.3.4",
		SSHUser:    "root",
		SSHPort:    "8022",
		MaxRetries: 3,
		RetryDelay: 0,
	}
	n := NewSSHNotifier(c)
	n.runCmd = func(name string, args ...string) error {
		calls.Add(1)
		return errFake
	}

	err := n.Send(Notification{Title: "T", Message: "M"})
	if err == nil {
		t.Error("expected error when all attempts fail")
	}
	if calls.Load() != 3 {
		t.Errorf("calls = %d, want 3", calls.Load())
	}
}

func TestWithRetry_ErrorMessage(t *testing.T) {
	c := &cfg.Config{
		AndroidIP:  "10.0.0.1",
		SSHUser:    "root",
		SSHPort:    "2222",
		MaxRetries: 2,
		RetryDelay: 0,
	}
	n := NewSSHNotifier(c)
	n.runCmd = func(name string, args ...string) error { return errFake }

	err := n.Send(Notification{Title: "T", Message: "M"})
	if err == nil {
		t.Fatal("expected error")
	}
	// Should mention attempts count and target
	if !strings.Contains(err.Error(), "failed after 2 attempts") {
		t.Errorf("error = %q, should mention attempts", err.Error())
	}
	if !strings.Contains(err.Error(), "10.0.0.1:2222") {
		t.Errorf("error = %q, should mention target", err.Error())
	}
}

func TestSendBackground_DoesNotBlock(t *testing.T) {
	c := &cfg.Config{
		AndroidIP:  "1.2.3.4",
		SSHUser:    "root",
		SSHPort:    "8022",
		MaxRetries: 1,
		RetryDelay: 0,
	}

	var calls atomic.Int32
	n := NewSSHNotifier(c)
	n.runCmd = func(name string, args ...string) error {
		calls.Add(1)
		return nil
	}

	n.SendBackground(Notification{Title: "T", Message: "M"})
	// Goroutine should eventually call runCmd
	time.Sleep(50 * time.Millisecond)
	if calls.Load() != 1 {
		t.Errorf("calls = %d, want 1", calls.Load())
	}
}

type fakeError struct{}

func (fakeError) Error() string { return "fake ssh error" }

var errFake error = fakeError{}
