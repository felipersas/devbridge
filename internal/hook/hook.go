package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/felipersas/devbridge/internal/cfg"
	"github.com/felipersas/devbridge/internal/notify"
	"github.com/felipersas/devbridge/internal/profile"
)

// Input is the JSON payload from Claude Code's Stop hook.
type Input struct {
	CWD                  string `json:"cwd"`
	LastAssistantMessage string `json:"last_assistant_message"`
}

const maxMessageLen = 200

// Run reads hook input from r and sends a notification (fire-and-forget).
func Run(r io.Reader) error {
	config, err := cfg.Load()
	if err != nil {
		return err
	}
	return RunWith(r, config, notify.NewSSHNotifier(config))
}

// RunWith accepts explicit dependencies for testing.
func RunWith(r io.Reader, config *cfg.Config, n notify.Notifier) error {
	var input Input
	if err := json.NewDecoder(r).Decode(&input); err != nil {
		return fmt.Errorf("parse hook input: %w", err)
	}

	cwd := input.CWD
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	dirName := filepath.Base(cwd)

	p := profile.Match(dirName)

	msg := input.LastAssistantMessage
	if msg == "" {
		msg = "Finished — needs your input"
	}
	if len(msg) > maxMessageLen {
		msg = msg[:maxMessageLen-3] + "..."
	}

	title := fmt.Sprintf("%s Claude Code — %s", p.Emoji, dirName)

	n.SendBackground(notify.Notification{
		Title:       title,
		Message:     msg,
		LEDColor:    p.LEDColor,
		Priority:    p.Priority,
		Group:       "claude-" + dirName,
		ID:          "claude-" + dirName,
		Sound:       config.Sound,
		TmuxSession: tmuxSession(),
	})
	n.Wait()

	return nil
}

// tmuxSession returns the current tmux session name, or empty if not in tmux.
var tmuxSession = func() string {
	if os.Getenv("TMUX") == "" {
		return ""
	}
	out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
