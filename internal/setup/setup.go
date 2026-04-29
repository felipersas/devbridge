package setup

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/felipersas/notifybridge/internal/cfg"
)

// Run starts the interactive setup wizard.
func Run() error {
	return RunWithReader(bufio.NewReader(os.Stdin))
}

// RunWithReader accepts an explicit reader for testing.
func RunWithReader(r *bufio.Reader) error {
	fmt.Println("NotifyBridge Setup Wizard")
	fmt.Println("=========================")
	fmt.Println()

	d := cfg.Default()

	androidIP := ask(r, "Android Tailscale IP", d.AndroidIP)
	sshUser := ask(r, "SSH user", d.SSHUser)
	sshPort := ask(r, "SSH port", d.SSHPort)
	title := ask(r, "Default notification title", d.DefaultTitle)
	sound := askBool(r, "Sound on notification", d.Sound)

	c := &cfg.Config{
		AndroidIP:    androidIP,
		SSHUser:      sshUser,
		SSHPort:      sshPort,
		DefaultTitle: title,
		Sound:        sound,
		Priority:     d.Priority,
		MaxRetries:   d.MaxRetries,
		RetryDelay:   d.RetryDelay,
	}

	path, err := cfg.Path()
	if err != nil {
		return err
	}

	if err := cfg.Write(path, c); err != nil {
		return err
	}

	fmt.Printf("\nConfig written to %s\n", path)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run scripts/termux-setup.sh on Android")
	fmt.Println("  2. Copy SSH key: ssh-copy-id -p PORT USER@IP")
	fmt.Println("  3. Test: notifybridge send 'Hello!'")
	return nil
}

func ask(r *bufio.Reader, prompt, def string) string {
	fmt.Printf("%s [%s]: ", prompt, def)
	input, _ := r.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return def
	}
	return input
}

func askBool(r *bufio.Reader, prompt string, def bool) bool {
	defStr := "y"
	if !def {
		defStr = "n"
	}
	fmt.Printf("%s (y/n) [%s]: ", prompt, defStr)
	input, _ := r.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return def
	}
	return input == "y" || input == "yes"
}
