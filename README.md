# Babything Agent

A lightweight bridge that forwards your local RTSP camera to the [Babything](https://github.com/ThatDavis/babything-cloud) cloud via WebRTC.

---

## Quick Start

### 1. Get an agent token

In your Babything admin dashboard, go to **Settings → Monitor** and click **"Generate Agent Token"**. Copy the token — you'll only see it once.

### 2. Download and run

#### Linux (headless binary + systemd)

Download the binary for your architecture from [GitHub Releases](https://github.com/ThatDavis/babything-agent/releases):

```bash
# AMD64 (most PCs and servers)
wget https://github.com/ThatDavis/babything-agent/releases/latest/download/babything-agent-linux-amd64

# ARM64 (Raspberry Pi 4/5, Apple Silicon VMs)
wget https://github.com/ThatDavis/babything-agent/releases/latest/download/babything-agent-linux-arm64

# ARMv7 (Raspberry Pi 3/Zero 2 W)
wget https://github.com/ThatDavis/babything-agent/releases/latest/download/babything-agent-linux-arm
```

Create a config file at `~/.config/babything/agent.yaml`:

```yaml
cloud_url: "wss://yourfamily.babything.app/monitor/agent"
rtsp_url: "rtsp://192.168.1.50:554/stream1"
agent_token: "your-token-here"
```

Run directly:
```bash
chmod +x babything-agent-linux-amd64
./babything-agent-linux-amd64
```

Or install as a systemd service:
```bash
sudo cp babything-agent-linux-amd64 /usr/local/bin/babything-agent
sudo mkdir -p /etc/babything
sudo cp your-config.yaml /etc/babything/agent.yaml
sudo cp build/linux/babything-agent.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now babything-agent
sudo journalctl -u babything-agent -f
```

#### Windows (desktop app)

1. Download `babything-agent.exe` from [GitHub Releases](https://github.com/ThatDavis/babything-agent/releases)
2. Double-click to run — it starts hidden in the background
3. Click the app icon in the taskbar to open settings
4. Fill in Cloud URL, RTSP URL, and Agent Token
5. Click **Save & Start**

#### macOS (desktop app)

1. Download `babything-agent-macos-universal.zip` from [GitHub Releases](https://github.com/ThatDavis/babything-agent/releases)
2. Extract and open `BabythingAgent.app`
3. The app runs in the background with a menu bar icon
4. Click the menu bar icon → **Open Settings**
5. Fill in Cloud URL, RTSP URL, and Agent Token
6. Click **Save & Start**

#### Docker (alternative)

```bash
docker run -d \
  --name babything-agent \
  --restart unless-stopped \
  -e CLOUD_URL=wss://yourfamily.babything.app/monitor/agent \
  -e RTSP_URL=rtsp://192.168.1.50:554/stream1 \
  -e AGENT_TOKEN=your-token-here \
  ghcr.io/babything/agent:latest
```

### 3. Open the Monitor tab

In your Babything web app, tap the **Monitor** tab. The agent will appear automatically and the stream will start.

---

## Requirements

| Requirement | Details |
|-------------|---------|
| **Camera** | RTSP-enabled IP camera (H.264 preferred) |
| **Network** | Agent must be on the same LAN as the camera |
| **Outbound** | HTTPS/WSS on port 443 to your Babything subdomain |
| **Hardware** | Raspberry Pi, old laptop, NAS, or any always-on machine |

### Finding your camera's RTSP URL

Common patterns by manufacturer:

| Brand | RTSP URL pattern |
|-------|------------------|
| Reolink | `rtsp://admin:password@192.168.1.100:554/h264Preview_01_main` |
| Amcrest | `rtsp://admin:password@192.168.1.100:554/cam/realmonitor?channel=1&subtype=0` |
| Hikvision | `rtsp://admin:password@192.168.1.100:554/Streaming/Channels/101` |
| Wyze (with firmware) | `rtsp://user:pass@192.168.1.100/live` |
| Generic | `rtsp://user:pass@ip:554/stream1` |

---

## How It Works

```
┌─────────────┐      WebSocket      ┌─────────────┐      WebRTC      ┌─────────────┐
│   Camera    │◄───── RTSP ───────►│    Agent    │◄──── P2P/TURN ──►│   Browser   │
│  (local)    │                     │  (local)    │                   │   (cloud)   │
└─────────────┘                     └──────┬──────┘                   └─────────────┘
                                           │
                                           │  Signaling
                                           │  (offer/answer/ICE)
                                           ▼
                                    ┌─────────────┐
                                    │  Cloud API  │
                                    │  /monitor   │
                                    └─────────────┘
```

1. **Agent connects** to your cloud tenant via secure WebSocket (`/monitor/agent`).
2. **Browser opens Monitor tab** and checks if an agent is connected.
3. **WebRTC handshake**: browser creates an SDP offer → cloud relays it to agent → agent creates answer → cloud relays it back.
4. **ICE candidate exchange** finds the best network path (direct P2P or TURN relay).
5. **Video flows** directly from agent to browser. The cloud only handles signaling — it never sees the video stream.

---

## Configuration

The agent can be configured via **config file** (preferred) or **environment variables**.

### Config file

Default paths:
- **Linux:** `~/.config/babything/agent.yaml`
- **Windows:** `%APPDATA%\Babything\agent.yaml`
- **macOS:** `~/Library/Application Support/Babything/agent.yaml`

Override with `CONFIG_PATH` env var.

Example `agent.yaml`:

```yaml
cloud_url: "wss://yourfamily.babything.app/monitor/agent"
rtsp_url: "rtsp://admin:pass@192.168.1.50:554/stream1"
agent_token: "eyJhbG..."

# Optional TURN server for NAT traversal
turn_url: ""
turn_username: ""
turn_password: ""
```

### Environment variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `CLOUD_URL` | **Yes** | WebSocket endpoint of your tenant | `wss://smith.babything.app/monitor/agent` |
| `RTSP_URL` | **Yes** | Full RTSP URL of your camera | `rtsp://admin:pass@192.168.1.50/stream` |
| `AGENT_TOKEN` | **Yes** | Agent token from admin dashboard | `eyJhbG...` |

---

## Building from Source

Requires **Go 1.23+**, **Node.js 20+**, and **ffmpeg**.

### Headless binary (all platforms)

```bash
cd babything-agent
go mod tidy
go build -o babything-agent ./cmd/agent
```

### Desktop app (Windows / macOS)

```bash
# Install Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build for current platform
wails build

# Build for specific platform
wails build -platform windows/amd64
wails build -platform darwin/universal
```

### Cross-compilation (Linux)

**Raspberry Pi (64-bit):**
```bash
GOOS=linux GOARCH=arm64 go build -o babything-agent-linux-arm64 ./cmd/agent
```

**Raspberry Pi (32-bit):**
```bash
GOOS=linux GOARCH=arm GOARM=7 go build -o babything-agent-linux-arm ./cmd/agent
```

### Docker build

```bash
docker build -t babything-agent .
```

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| "Agent disconnected" in Monitor tab | Agent can't reach cloud | Check `cloud_url` and outbound HTTPS/WSS |
| "Connection timed out" | WebRTC signaling failed | Check that agent is running and token is valid |
| Black screen, no video | Camera codec not H.264 | Verify camera outputs H.264, or ffmpeg will transcode |
| Choppy / high CPU | ffmpeg transcoding | Lower camera resolution or ensure `-c:v copy` works |
| Works on WiFi, not on cellular | NAT / firewall blocking P2P | Enable TURN server in cloud deployment |

### Check agent logs

**systemd (Linux):**
```bash
sudo journalctl -u babything-agent -f
```

**Desktop app (Windows / macOS):**
Logs are printed to stdout. On macOS, run from Terminal:
```bash
/Applications/BabythingAgent.app/Contents/MacOS/babything-agent
```

**Docker:**
```bash
docker logs babything-agent
```

### Test RTSP directly

```bash
ffplay -rtsp_transport tcp rtsp://your-camera-url
```

If this doesn't show video, the agent won't either.

---

## Security

- The agent token is a **JWT** scoped only to monitor access. It expires in 30 days.
- All signaling is over **WSS** (WebSocket Secure).
- Video never transits through cloud servers — it flows **directly** from your agent to your browser.
- Revoke a token by generating a new one in the admin dashboard.

---

## License

Same as Babything — open source under the project's chosen license.
