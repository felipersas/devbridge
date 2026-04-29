package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/felipersas/notifybridge/internal/cfg"
	"github.com/felipersas/notifybridge/internal/notify"
	"github.com/felipersas/notifybridge/internal/profile"
)

// Input is the JSON payload from Claude Code's Stop hook.
type Input struct {
	CWD                  string `json:"cwd"`
	LastAssistantMessage string `json:"last_assistant_message"`
}

const maxMessageLen = 200

// Run reads hook input from r and sends a notification (fire-and-forget).
func Run(r io.Reader) error {
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

	config, err := cfg.Load()
	if err != nil {
		return err
	}

	notifier := notify.NewSSHNotifier(config)
	notifier.SendBackground(notify.Notification{
		Title:    title,
		Message:  msg,
		LEDColor: p.LEDColor,
		Priority: p.Priority,
		Group:    "claude-" + dirName,
		ID:       "claude-" + dirName,
		Sound:    config.Sound,
	})

	return nil
}
