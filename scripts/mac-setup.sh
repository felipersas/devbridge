#!/usr/bin/env bash
# NotifyBridge - macOS Setup
# Run this once on your Mac

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

info() { echo -e "${CYAN}[INFO]${NC} $1"; }
ok()   { echo -e "${GREEN}[OK]${NC} $1"; }
err()  { echo -e "${RED}[ERROR]${NC} $1"; }

echo ""
echo "=============================="
echo "  NotifyBridge - Mac Setup"
echo "=============================="
echo ""

# 1. Check SSH key
if [ ! -f ~/.ssh/id_ed25519 ]; then
    info "Generating SSH key..."
    ssh-keygen -t ed25519 -C "notifybridge@mac" -f ~/.ssh/id_ed25519 -N ""
    ok "SSH key generated"
else
    ok "SSH key exists at ~/.ssh/id_ed25519"
fi

# 2. Show public key
echo ""
info "Your public key (paste this into Termux setup):"
echo ""
cat ~/.ssh/id_ed25519.pub
echo ""

# 3. Detect Android Tailscale IP
ANDROID_IP=""
if command -v tailscale &>/dev/null; then
    ANDROID_IP=$(tailscale status 2>/dev/null | grep android | awk '{print $1}' | head -1)
fi

# 4. Create config
CONF="$HOME/.notifybridge.conf"
if [ -f "$CONF" ]; then
    ok "Config exists at $CONF"
else
    cp "$SCRIPT_DIR/notifybridge.conf.example" "$CONF"

    if [ -n "$ANDROID_IP" ]; then
        sed -i '' "s/ANDROID_IP=\"100.115.83.120\"/ANDROID_IP=\"$ANDROID_IP\"/" "$CONF"
        ok "Android IP auto-detected: $ANDROID_IP"
    fi

    ok "Config created at $CONF"
fi

# 5. Install notify script
DEST="/usr/local/bin/notify"
if [ -w /usr/local/bin ]; then
    cp "$SCRIPT_DIR/notify" "$DEST"
    chmod +x "$DEST"
    ok "notify script installed at $DEST"
else
    info "Need sudo to install to /usr/local/bin"
    sudo cp "$SCRIPT_DIR/notify" "$DEST"
    sudo chmod +x "$DEST"
    ok "notify script installed at $DEST"
fi

# 6. Test connection
echo ""
info "Testing SSH connection to Android..."
if ssh -o ConnectTimeout=5 \
       -o BatchMode=yes \
       -p 8022 \
       "root@${ANDROID_IP:-100.115.83.120}" \
       "echo OK" 2>/dev/null; then
    ok "SSH connection works!"
    echo ""
    info "Sending test notification..."
    notify "NotifyBridge is working!"
    ok "Check your Android device for the notification"
else
    err "SSH connection failed. Make sure:"
    echo "  1. Termux sshd is running (run 'sshd' in Termux)"
    echo "  2. Your public key is in ~/.ssh/authorized_keys on Termux"
    echo "  3. Tailscale is connected on both devices"
    echo ""
    echo "  Run termux-setup.sh on your Android device first"
    echo "  Then paste the public key shown above"
fi

echo ""
echo "=============================="
ok "Mac setup complete!"
echo ""
echo "Usage:"
echo "  notify 'Build finished'"
echo "  npm run build && notify 'Success' || notify 'Failed'"
echo ""
