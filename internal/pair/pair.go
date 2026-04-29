package pair

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/mdp/qrterminal"

	"github.com/felipersas/notifybridge/internal/cfg"
)

const defaultPort = 19876

// PhoneInfo holds phone connection details discovered during pairing.
type PhoneInfo struct {
	IP         string `json:"ip"`
	SSHPort    string `json:"ssh_port"`
	User       string `json:"user"`
	DeviceName string `json:"device_name"`
}

// Run executes the full pairing flow with a 5-minute timeout.
func Run() error {
	return RunWithTimeout(5 * time.Minute)
}

// RunWithTimeout starts the pairing server, shows a QR code, and waits
// for a phone to register. On success it writes the config file.
func RunWithTimeout(timeout time.Duration) error {
	ip, err := reachableIP()
	if err != nil {
		return fmt.Errorf("no reachable IP: %w", err)
	}

	pubKey, err := ensureSSHPubKey()
	if err != nil {
		return fmt.Errorf("SSH key: %w", err)
	}

	if err := ensureTmux(); err != nil {
		return fmt.Errorf("tmux: %w", err)
	}

	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("token: %w", err)
	}

	url := fmt.Sprintf("http://%s:%d/%s", ip, defaultPort, token)

	resultCh := make(chan PhoneInfo, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/"+token, handleIndex(ip, token))
	mux.HandleFunc("/"+token+"/setup.sh", handleSetup(ip, token, resultCh, pubKey))
	mux.HandleFunc("/"+token+"/register", handleRegister(ip, resultCh, pubKey))

	srv := &http.Server{Addr: fmt.Sprintf(":%d", defaultPort), Handler: mux}
	go srv.ListenAndServe()

	showQR(url, token)

	select {
	case phone := <-resultCh:
		srv.Close()
		if err := writeConfig(&phone); err != nil {
			return err
		}
		fmt.Println()
		fmt.Printf("  Paired with %s (%s)\n", phone.DeviceName, phone.IP)
		path, _ := cfg.Path()
		fmt.Printf("  Config: %s\n", path)
		fmt.Println("  Test: notifybridge send \"Hello!\"")
		return nil

	case <-time.After(timeout):
		srv.Close()
		return fmt.Errorf("pairing timed out after %v", timeout)
	}
}

func showQR(url, token string) {
	fmt.Println()
	fmt.Println("  NotifyBridge Pairing")
	fmt.Println("  ═══════════════════")
	fmt.Println()
	qrterminal.Generate(url, qrterminal.L, os.Stdout)
	fmt.Println()
	fmt.Printf("  Code: %s\n", token)
	fmt.Printf("  URL:  %s\n", url)
	fmt.Println()
	fmt.Println("  1. Scan QR on phone (or open URL)")
	fmt.Println("  2. Run the Termux command shown in browser")
	fmt.Println("  3. Wait for pairing to complete...")
	fmt.Println()
}

func writeConfig(phone *PhoneInfo) error {
	c := &cfg.Config{
		AndroidIP:    phone.IP,
		SSHUser:      phone.User,
		SSHPort:      phone.SSHPort,
		DefaultTitle: "Mac Remote",
		Sound:        true,
		Priority:     "high",
		MaxRetries:   2,
		RetryDelay:   3,
	}
	path, err := cfg.Path()
	if err != nil {
		return err
	}
	return cfg.Write(path, c)
}

// --- HTTP Handlers ---

func handleIndex(ip, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html><head>
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>NotifyBridge Pairing</title>
<style>
body{font-family:system-ui,sans-serif;max-width:500px;margin:40px auto;padding:0 20px;background:#1a1a2e;color:#e0e0e0}
h1{color:#00d4ff}
.cmd-wrap{position:relative;background:#16213e;border-radius:8px;overflow:hidden}
pre{padding:16px;overflow-x:auto;color:#00ff88;font-size:14px;margin:0}
.copy-btn{position:absolute;top:8px;right:8px;background:#00d4ff;color:#1a1a2e;border:none;padding:6px 14px;border-radius:6px;font-size:13px;font-weight:600;cursor:pointer;transition:all .2s}
.copy-btn:hover{background:#00b8d4}
.copy-btn.copied{background:#00e676}
.note{color:#888;font-size:0.85em;margin-top:20px}
</style>
</head><body>
<h1>NotifyBridge Pairing</h1>
<p>Run this command in <strong>Termux</strong>:</p>
<div class="cmd-wrap">
<pre id="cmd">curl -sS http://%s:%d/%s/setup.sh | bash</pre>
<button class="copy-btn" onclick="copyCmd()">Copy</button>
</div>
<p class="note">Make sure Tailscale is running on this phone for remote notifications.</p>
<script>
function copyCmd(){var t=document.getElementById("cmd").textContent;navigator.clipboard.writeText(t).then(function(){var b=document.querySelector(".copy-btn");b.textContent="Copied!";b.classList.add("copied");setTimeout(function(){b.textContent="Copy";b.classList.remove("copied")},2000)})}
</script>
</body></html>`, ip, defaultPort, token)
	}
}

func handleSetup(ip, token string, resultCh chan<- PhoneInfo, pubKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/x-shellscript")
		fmt.Fprintf(w, `#!/data/data/com.termux/files/usr/bin/bash
# NotifyBridge Phone Setup
set -e

echo ""
echo "  NotifyBridge Phone Setup"
echo "  ════════════════════════"
echo ""

# Install packages
echo "  Installing packages..."
pkg install -y openssh termux-api 2>/dev/null

# Generate SSH key if not exists
if [ ! -f ~/.ssh/id_ed25519 ]; then
    echo "  Generating SSH key..."
    mkdir -p ~/.ssh
    ssh-keygen -t ed25519 -N "" -f ~/.ssh/id_ed25519 -q
fi

# Start sshd if not running
if ! pgrep sshd >/dev/null 2>&1; then
    echo "  Starting SSH server..."
    sshd
fi

# Enable wake lock so Termux keeps running
termux-wake-lock 2>/dev/null || true

# Get device info
DEVICE=$(getprop ro.product.model 2>/dev/null || hostname)
USER=$(whoami)
PUBKEY=$(cat ~/.ssh/id_ed25519.pub)

# Register with PC (send our pubkey for reverse SSH)
echo "  Registering with PC..."
RESPONSE=$(curl -sS -X POST "http://%s:%d/%s/register" \
    -H "Content-Type: application/json" \
    -d "{\"ssh_port\":\"8022\",\"user\":\"$USER\",\"device_name\":\"$DEVICE\",\"pubkey\":\"$PUBKEY\"}")

# Extract PC pubkey from response and add to authorized_keys
PC_PUBKEY=$(echo "$RESPONSE" | sed -n 's/.*"pc_pubkey":"\([^"]*\)".*/\1/p')
if [ -n "$PC_PUBKEY" ]; then
    mkdir -p ~/.ssh
    echo "$PC_PUBKEY" >> ~/.ssh/authorized_keys
    sort -u ~/.ssh/authorized_keys -o ~/.ssh/authorized_keys 2>/dev/null
    chmod 600 ~/.ssh/authorized_keys
fi

# Save Mac connection info for remote Claude Code
MAC_IP=$(echo "$RESPONSE" | sed -n 's/.*"mac_ip":"\([^"]*\)".*/\1/p')
MAC_SSH_PORT=$(echo "$RESPONSE" | sed -n 's/.*"mac_ssh_port":"\([^"]*\)".*/\1/p')
MAC_USER=$(echo "$RESPONSE" | sed -n 's/.*"mac_user":"\([^"]*\)".*/\1/p')

if [ -n "$MAC_IP" ]; then
    echo "  Saving remote connection..."
    cat > ~/.notifybridge-remote.conf <<REMOTE_CONF
MAC_IP=$MAC_IP
MAC_SSH_PORT=$MAC_SSH_PORT
MAC_USER=$MAC_USER
REMOTE_CONF
fi

# Install claude-remote command
echo "  Installing claude-remote..."
cat > $PREFIX/bin/claude-remote <<'REMOTE_SCRIPT'
#!/data/data/com.termux/files/usr/bin/bash
# Connect to Claude Code on Mac via SSH+tmux
source ~/.notifybridge-remote.conf 2>/dev/null || { echo "Run notifybridge pair first"; exit 1; }

ssh -t -o ConnectTimeout=5 -o StrictHostKeyChecking=accept-new \
    -p "${MAC_SSH_PORT:-22}" "${MAC_USER:-$USER}@${MAC_IP}" \
    "bash -l -c 'tmux new-session -A -s claude claude'"
REMOTE_SCRIPT
chmod +x $PREFIX/bin/claude-remote

echo ""
echo "  Setup complete!"
echo "  Remote access: claude-remote"
echo "  Return to your PC - pairing should be confirmed."
echo ""
`, ip, defaultPort, token)
	}
}

func handleRegister(macIP string, resultCh chan<- PhoneInfo, pubKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			SSHPort    string `json:"ssh_port"`
			User       string `json:"user"`
			DeviceName string `json:"device_name"`
			PubKey     string `json:"pubkey"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		phoneIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			phoneIP = r.RemoteAddr
		}
		if req.SSHPort == "" {
			req.SSHPort = "8022"
		}
		if req.User == "" {
			req.User = "root"
		}

		// Add phone's pubkey to Mac's authorized_keys
		if req.PubKey != "" {
			if err := addAuthorizedKey(req.PubKey); err != nil {
				http.Error(w, fmt.Sprintf("add key: %v", err), http.StatusInternalServerError)
				return
			}
		}

		macUser, _ := osuserHome()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":       "ok",
			"pc_pubkey":    pubKey,
			"mac_ip":       macIP,
			"mac_ssh_port": "22",
			"mac_user":     macUser,
		})

		select {
		case resultCh <- PhoneInfo{
			IP:         phoneIP,
			SSHPort:    req.SSHPort,
			User:       req.User,
			DeviceName: req.DeviceName,
		}:
		default:
		}
	}
}

// --- Helpers ---

func generateToken() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n), nil
}

func ensureTmux() error {
	if _, err := exec.LookPath("tmux"); err == nil {
		return nil
	}
	fmt.Println("  Installing tmux...")
	cmd := exec.Command("brew", "install", "tmux")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func reachableIP() (string, error) {
	// Prefer Tailscale
	if out, err := exec.Command("tailscale", "ip", "-4").Output(); err == nil {
		if ip := strings.TrimSpace(string(out)); ip != "" {
			return ip, nil
		}
	}

	// Fallback: any non-loopback IPv4 (prefer 100.x range)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			if strings.HasPrefix(ipnet.IP.String(), "100.") {
				return ipnet.IP.String(), nil
			}
		}
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("no reachable IP found")
}

func ensureSSHPubKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Use existing key
	for _, name := range []string{"id_ed25519", "id_rsa"} {
		data, err := os.ReadFile(home + "/.ssh/" + name + ".pub")
		if err == nil {
			return strings.TrimSpace(string(data)), nil
		}
	}

	// Generate a dedicated key
	keyPath := home + "/.ssh/notifybridge_ed25519"
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-N", "", "-f", keyPath, "-q")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ssh-keygen: %w", err)
	}
	data, err := os.ReadFile(keyPath + ".pub")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func osuserHome() (string, string) {
	u, err := user.Current()
	if err != nil {
		return "root", ""
	}
	return u.Username, u.HomeDir
}

func addAuthorizedKey(pubKey string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	sshDir := home + "/.ssh"
	os.MkdirAll(sshDir, 0700)

	authFile := sshDir + "/authorized_keys"

	// Read existing keys
	data, _ := os.ReadFile(authFile)
	existing := strings.TrimSpace(string(data))

	pubKey = strings.TrimSpace(pubKey)

	// Check if key already exists
	for _, line := range strings.Split(existing, "\n") {
		if strings.TrimSpace(line) == pubKey {
			return nil // already present
		}
	}

	// Append new key
	f, err := os.OpenFile(authFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if existing != "" && !strings.HasSuffix(existing, "\n") {
		f.WriteString("\n")
	}
	f.WriteString(pubKey + "\n")
	return nil
}
