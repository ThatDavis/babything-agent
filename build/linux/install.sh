#!/bin/bash
set -e

BINARY="babything-agent"
SERVICE="babything-agent.service"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/babything"

echo "Installing Babything Agent..."

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  BINARY_ARCH="babything-agent-linux-amd64" ;;
    aarch64) BINARY_ARCH="babything-agent-linux-arm64" ;;
    armv7l)  BINARY_ARCH="babything-agent-linux-arm" ;;
    *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

if [ ! -f "$BINARY_ARCH" ]; then
    echo "Binary not found: $BINARY_ARCH"
    echo "Please download the correct binary for your architecture."
    exit 1
fi

# Install binary
echo "Installing binary to $INSTALL_DIR..."
cp "$BINARY_ARCH" "$INSTALL_DIR/$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"

# Create config directory
echo "Creating config directory $CONFIG_DIR..."
mkdir -p "$CONFIG_DIR"

# Install systemd service if systemd is available
if command -v systemctl &> /dev/null; then
    echo "Installing systemd service..."
    cp "$SERVICE" /etc/systemd/system/
    systemctl daemon-reload
    systemctl enable "$SERVICE"
    echo "Service installed. Start with: sudo systemctl start $SERVICE"
else
    echo "systemd not found. Service file installed to $(pwd)/$SERVICE"
fi

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "  1. Create config at $CONFIG_DIR/agent.yaml"
echo "  2. Start with: sudo systemctl start $SERVICE"
echo "  3. View logs with: sudo journalctl -u $SERVICE -f"
