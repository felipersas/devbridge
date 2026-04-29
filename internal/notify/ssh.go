package notify

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/felipersas/devbridge/internal/cfg"
)

// SSHNotifier sends notifications via SSH to Termux.
type SSHNotifier struct {
	config *cfg.Config
	runCmd func(name string, args ...string) error
	wg     sync.WaitGroup
}

// NewSSHNotifier creates a new SSH-based notifier.
func NewSSHNotifier(c *cfg.Config) *SSHNotifier {
	return &SSHNotifier{config: c, runCmd: defaultRunCmd}
}

func defaultRunCmd(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

// Send delivers a notification synchronously with retries.
func (s *SSHNotifier) Send(n Notification) error {
	cmdStr := buildTermuxCommand(n)
	return s.withRetry(cmdStr)
}

// SendBackground delivers a notification in a goroutine (use Wait to block).
func (s *SSHNotifier) SendBackground(n Notification) {
	cmdStr := buildTermuxCommand(n)
	args := s.sshArgs(cmdStr)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		_ = s.runCmd("ssh", args...)
	}()
}

// Wait blocks until all background sends complete.
func (s *SSHNotifier) Wait() {
	s.wg.Wait()
}

func (s *SSHNotifier) sshArgs(cmdStr string) []string {
	addr := fmt.Sprintf("%s@%s", s.config.SSHUser, s.config.AndroidIP)
	return []string{
		"-o", "ConnectTimeout=5",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
		"-p", s.config.SSHPort,
		addr,
		cmdStr,
	}
}

func (s *SSHNotifier) withRetry(cmdStr string) error {
	args := s.sshArgs(cmdStr)

	var lastErr error
	for attempt := 0; attempt < s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(s.config.RetryDelay) * time.Second)
		}

		if err := s.runCmd("ssh", args...); err != nil {
			lastErr = err
			continue
		}
		return nil
	}

	return fmt.Errorf("failed after %d attempts to %s:%s: %w",
		s.config.MaxRetries, s.config.AndroidIP, s.config.SSHPort, lastErr)
}

func buildTermuxCommand(n Notification) string {
	parts := []string{"termux-notification"}
	parts = append(parts, "--title", shellEscape(n.Title))
	parts = append(parts, "--content", shellEscape(n.Message))
	if n.Sound {
		parts = append(parts, "--sound")
	}
	if n.Priority != "" {
		parts = append(parts, "--priority", n.Priority)
	}
	if n.LEDColor != "" {
		parts = append(parts, "--led-color", n.LEDColor)
	}
	if n.Group != "" {
		parts = append(parts, "--group", shellEscape(n.Group))
	}
	if n.ID != "" {
		parts = append(parts, "--id", shellEscape(n.ID))
	}
	if n.TmuxSession != "" {
		action := "claude-remote -s " + shellEscape(n.TmuxSession)
		parts = append(parts, "--button1", shellEscape("Open Session"))
		parts = append(parts, "--button1-action", shellEscape(action))
	}
	return strings.Join(parts, " ")
}

// shellEscape wraps a string in single quotes, escaping embedded single quotes.
func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
