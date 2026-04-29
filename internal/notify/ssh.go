package notify

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/felipersas/notifybridge/internal/cfg"
)

// SSHNotifier sends notifications via SSH to Termux.
type SSHNotifier struct {
	config *cfg.Config
}

// NewSSHNotifier creates a new SSH-based notifier.
func NewSSHNotifier(c *cfg.Config) *SSHNotifier {
	return &SSHNotifier{config: c}
}

// Send delivers a notification synchronously with retries.
func (s *SSHNotifier) Send(n Notification) error {
	cmdStr := buildTermuxCommand(n)
	return s.withRetry(cmdStr)
}

// SendBackground delivers a notification without waiting (fire-and-forget).
func (s *SSHNotifier) SendBackground(n Notification) {
	cmdStr := buildTermuxCommand(n)
	addr := fmt.Sprintf("%s@%s", s.config.SSHUser, s.config.AndroidIP)
	cmd := exec.Command("ssh",
		"-o", "ConnectTimeout=5",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
		"-p", s.config.SSHPort,
		addr,
		cmdStr,
	)
	_ = cmd.Start()
}

func (s *SSHNotifier) withRetry(cmdStr string) error {
	addr := fmt.Sprintf("%s@%s", s.config.SSHUser, s.config.AndroidIP)

	var lastErr error
	for attempt := 0; attempt < s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(s.config.RetryDelay) * time.Second)
		}

		cmd := exec.Command("ssh",
			"-o", "ConnectTimeout=5",
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "BatchMode=yes",
			"-p", s.config.SSHPort,
			addr,
			cmdStr,
		)

		if err := cmd.Run(); err != nil {
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
	return strings.Join(parts, " ")
}

// shellEscape wraps a string in single quotes, escaping embedded single quotes.
func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
