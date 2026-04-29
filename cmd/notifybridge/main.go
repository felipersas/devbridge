package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/felipersas/notifybridge/internal/cfg"
	"github.com/felipersas/notifybridge/internal/health"
	"github.com/felipersas/notifybridge/internal/hook"
	"github.com/felipersas/notifybridge/internal/notify"
	"github.com/felipersas/notifybridge/internal/pair"
	"github.com/felipersas/notifybridge/internal/setup"
)

var version = "dev"

var (
	flagTitle    string
	flagLEDColor string
	flagGroup    string
	flagID       string
	flagPriority string
	flagSound    bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "notifybridge",
		Short: "Send notifications from Mac to Android via SSH/Tailscale",
	}

	rootCmd.AddCommand(sendCmd(), hookCmd(), pairCmd(), setupCmd(), testCmd(), configCmd(), versionCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func sendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [message]",
		Short: "Send a notification",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := cfg.Load()
			if err != nil {
				return err
			}

			title := flagTitle
			if title == "" {
				title = config.DefaultTitle
			}

			notifier := notify.NewSSHNotifier(config)
			return notifier.Send(notify.Notification{
				Title:    title,
				Message:  args[0],
				LEDColor: flagLEDColor,
				Group:    flagGroup,
				ID:       flagID,
				Priority: flagPriority,
				Sound:    flagSound,
			})
		},
	}

	cmd.Flags().StringVarP(&flagTitle, "title", "t", "", "Notification title")
	cmd.Flags().StringVar(&flagLEDColor, "led-color", "", "LED color hex (e.g. FF0000)")
	cmd.Flags().StringVar(&flagGroup, "group", "", "Notification group key")
	cmd.Flags().StringVar(&flagID, "id", "", "Notification channel ID")
	cmd.Flags().StringVar(&flagPriority, "priority", "", "high|default|low|max|min")
	cmd.Flags().BoolVar(&flagSound, "sound", true, "Play notification sound")

	return cmd
}

func hookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Claude Code hook (reads JSON from stdin)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return hook.Run(os.Stdin)
		},
	}
}

func pairCmd() *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "pair",
		Short: "Pair with an Android phone via QR code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pair.RunWithTimeout(timeout)
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "Pairing timeout")
	return cmd
}

func setupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setup.Run()
		},
	}
}

func testCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test connectivity to Android device",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := cfg.Load()
			if err != nil {
				return err
			}

			fmt.Println("NotifyBridge Connectivity Test")
			fmt.Println("==============================")
			fmt.Println()

			result := health.Check(config, nil)

			if result.SSHConnected {
				fmt.Println("  SSH:    connected")
			} else {
				fmt.Printf("  SSH:    FAILED (%s)\n", result.ErrorMessage)
			}

			if result.TermuxAvailable {
				fmt.Println("  Termux: available")
			} else if result.SSHConnected {
				fmt.Println("  Termux: unavailable (install termux-api)")
			}

			fmt.Println()
			if !result.SSHConnected || !result.TermuxAvailable {
				return fmt.Errorf("connectivity check failed")
			}
			fmt.Println("  All checks passed!")
			return nil
		},
	}
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := cfg.Load()
			if err != nil {
				return err
			}
			path, _ := cfg.Path()
			fmt.Printf("Config: %s\n\n", path)
			fmt.Printf("  Android IP:    %s\n", config.AndroidIP)
			fmt.Printf("  SSH User:      %s\n", config.SSHUser)
			fmt.Printf("  SSH Port:      %s\n", config.SSHPort)
			fmt.Printf("  Default Title: %s\n", config.DefaultTitle)
			fmt.Printf("  Sound:         %v\n", config.Sound)
			fmt.Printf("  Priority:      %s\n", config.Priority)
			fmt.Printf("  Max Retries:   %d\n", config.MaxRetries)
			fmt.Printf("  Retry Delay:   %ds\n", config.RetryDelay)
			return nil
		},
	})

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("notifybridge %s\n", version)
		},
	}
}
