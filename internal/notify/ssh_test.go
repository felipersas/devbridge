package notify

import (
	"strings"
	"testing"
)

func TestShellEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "'hello'"},
		{"it's a test", "'it'\\''s a test'"},
		{"simple", "'simple'"},
		{"", "''"},
		{"hello world", "'hello world'"},
	}

	for _, tt := range tests {
		got := shellEscape(tt.input)
		if got != tt.want {
			t.Errorf("shellEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBuildTermuxCommand(t *testing.T) {
	n := Notification{
		Title:    "Test Title",
		Message:  "Hello World",
		Priority: "high",
		Sound:    true,
	}

	cmd := buildTermuxCommand(n)

	if !strings.Contains(cmd, "termux-notification") {
		t.Error("command should contain termux-notification")
	}
	if !strings.Contains(cmd, "--title 'Test Title'") {
		t.Errorf("command should contain --title, got: %s", cmd)
	}
	if !strings.Contains(cmd, "--content 'Hello World'") {
		t.Errorf("command should contain --content, got: %s", cmd)
	}
	if !strings.Contains(cmd, "--sound") {
		t.Error("command should contain --sound")
	}
	if !strings.Contains(cmd, "--priority high") {
		t.Errorf("command should contain --priority, got: %s", cmd)
	}
}

func TestBuildTermuxCommandNoSound(t *testing.T) {
	n := Notification{
		Title:   "Test",
		Message: "No sound",
		Sound:   false,
	}

	cmd := buildTermuxCommand(n)

	if strings.Contains(cmd, "--sound") {
		t.Error("command should not contain --sound when Sound=false")
	}
}

func TestBuildTermuxCommandWithOptional(t *testing.T) {
	n := Notification{
		Title:    "Test",
		Message:  "Msg",
		LEDColor: "FF0000",
		Group:    "build",
		ID:       "build-1",
		Priority: "low",
		Sound:    true,
	}

	cmd := buildTermuxCommand(n)

	if !strings.Contains(cmd, "--led-color FF0000") {
		t.Errorf("missing --led-color, got: %s", cmd)
	}
	if !strings.Contains(cmd, "--group 'build'") {
		t.Errorf("missing --group, got: %s", cmd)
	}
	if !strings.Contains(cmd, "--id 'build-1'") {
		t.Errorf("missing --id, got: %s", cmd)
	}
}
