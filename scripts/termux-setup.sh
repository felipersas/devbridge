#!/data/data/com.termux/files/usr/bin/bash
# NotifyBridge - Termux (Android) Setup
# Run this ONCE on your Android/Termux device

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { echo -e "${CYAN}[INFO]${NC} $1"; }
ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
err()   { echo -e "${RED}[ERROR]${NC} $1"; }

echo ""
echo "=============================="
echo "  NotifyBridge - Termux Setup"
echo "=============================="
echo ""

# 1. Update packages
info "Updating packages..."
pkg update -y && pkg upgrade -y

# 2. Install dependencies
info "Installing dependencies..."
pkg install -y openssh termux-api

# 3. Check Termux:API app
if ! command -v termux-notification &>/dev/null; then
    err "termux-notification not found."
    echo "  Install 'Termux:API' app from F-Droid or:"
    echo "  https://f-droid.org/packages/com.termux.api/"
    exit 1
fi
ok "termux-api CLI available"

# 4. Generate host key if needed
if [ ! -f ~/.ssh/ssh_host_ed25519_key ]; then
    info "Generating SSH host key..."
    ssh-keygen -t ed25519 -f ~/.ssh/ssh_host_ed25519_key -N ""
    ok "Host key generated"
else
    ok "Host key already exists"
fi

# 5. Create .ssh directory
mkdir -p ~/.ssh
chmod 700 ~/.ssh
touch ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys

# 6. Ask for Mac's public key
echo ""
warn "Paste your Mac's PUBLIC key (from: cat ~/.ssh/id_ed25519.pub)"
echo "  (Press Enter on empty line when done)"
echo ""

KEY=""
while IFS= read -r line; do
    [ -z "$line" ] && break
    KEY="$line"
done

if [ -n "$KEY" ]; then
    # Avoid duplicates
    if grep -qF "$KEY" ~/.ssh/authorized_keys 2>/dev/null; then
        ok "Key already in authorized_keys"
    else
        echo "$KEY" >> ~/.ssh/authorized_keys
        ok "Key added to authorized_keys"
    fi
else
    warn "No key pasted. Add manually later:"
    warn "  echo 'YOUR_KEY' >> ~/.ssh/authorized_keys"
fi

# 7. Configure sshd port
SSHD_CONFIG="$PREFIX/etc/ssh/sshd_config"
if [ -f "$SSHD_CONFIG" ]; then
    if grep -q "^Port " "$SSHD_CONFIG"; then
        sed -i 's/^Port .*/Port 8022/' "$SSHD_CONFIG"
    else
        echo "Port 8022" >> "$SSHD_CONFIG"
    fi
    ok "sshd configured on port 8022"
fi

# 8. Start sshd
info "Starting sshd..."
if pgrep -f "sshd" >/dev/null; then
    ok "sshd already running"
else
    sshd
    ok "sshd started on port 8022"
fi

# 9. Wake lock
info "Enabling wake lock..."
termux-wake-lock
ok "Wake lock active (SSH stays alive with screen off)"

# 10. Boot script
BOOT_SCRIPT=~/.termux/boot/notifybridge-sshd.sh
mkdir -p ~/.termux/boot
cat > "$BOOT_SCRIPT" << 'BOOT'
#!/data/data/com.termux/files/usr/bin/bash
# Auto-start sshd + wake lock on boot
termux-wake-lock
sshd
BOOT
chmod +x "$BOOT_SCRIPT"
ok "Boot script created at $BOOT_SCRIPT"

# 11. Get Tailscale IP
echo ""
info "Your Tailscale IP:"
if command -v tailscale &>/dev/null; then
    tailscale ip -4 2>/dev/null || echo "  (Tailscale not running on Termux)"
else
    echo "  Install Tailscale on Android to get your IP"
    echo "  Or check: https://login.tailscale.com/admin/machines"
fi

echo ""
echo "=============================="
ok "Termux setup complete!"
echo ""
echo "Next steps on Mac:"
echo "  1. cp notifybridge.conf.example ~/.notifybridge.conf"
echo "  2. Edit ANDROID_IP in ~/.notifybridge.conf"
echo "  3. Run: notify 'Hello from Mac!'"
echo ""
