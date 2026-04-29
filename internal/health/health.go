package health

import (
	"fmt"
	"os/exec"

	"github.com/felipersas/notifybridge/internal/cfg"
)

// Result holds the outcome of a connectivity check.
type Result struct {
	SSHConnected    bool
	TermuxAvailable bool
	ErrorMessage    string
}

// Check tests connectivity to the configured Android device.
// If execFn is nil, it uses real exec.Command.
func Check(config *cfg.Config, execFn func(name string, args ...string) error) *Result {
	if execFn == nil {
		execFn = func(name string, args ...string) error {
			return exec.Command(name, args...).Run()
		}
	}

	r := &Result{}
	addr := fmt.Sprintf("%s@%s", config.SSHUser, config.AndroidIP)

	// Test basic SSH connectivity
	sshArgs := []string{
		"-o", "ConnectTimeout=5",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-p", config.SSHPort,
		addr,
		"echo", "OK",
	}
	if err := execFn("ssh", sshArgs...); err != nil {
		r.ErrorMessage = err.Error()
		return r
	}
	r.SSHConnected = true

	// Test termux-notification availability
	termuxArgs := []string{
		"-o", "ConnectTimeout=5",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-p", config.SSHPort,
		addr,
		"termux-notification --title 'NotifyBridge Test' --content 'Connectivity OK'",
	}
	if err := execFn("ssh", termuxArgs...); err == nil {
		r.TermuxAvailable = true
	}

	return r
}
