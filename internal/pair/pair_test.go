package pair

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(token) != 6 {
		t.Errorf("token length = %d, want 6", len(token))
	}
	for _, c := range token {
		if c < '0' || c > '9' {
			t.Errorf("token contains non-digit: %c", c)
		}
	}
}

func TestGenerateTokenUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateToken()
		if err != nil {
			t.Fatal(err)
		}
		if seen[token] {
			t.Errorf("duplicate token: %s", token)
		}
		seen[token] = true
	}
}

func TestReachableIP(t *testing.T) {
	ip, err := reachableIP()
	if err != nil {
		t.Skipf("no reachable IP: %v", err)
	}
	if ip == "" {
		t.Error("expected non-empty IP")
	}
}

func TestEnsureSSHPubKey(t *testing.T) {
	pubKey, err := ensureSSHPubKey()
	if err != nil {
		t.Skipf("SSH key not available: %v", err)
	}
	if pubKey == "" {
		t.Error("expected non-empty public key")
	}
	if len(pubKey) < 50 {
		t.Errorf("pubkey seems too short: %s", pubKey)
	}
}
