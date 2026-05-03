# Babything Agent

A lightweight bridge that forwards your local RTSP camera to the [Babything](https://github.com/ThatDavis/babything-cloud) cloud via WebRTC.

---

## Quick Start

### 1. Get an agent token

In your Babything admin dashboard, go to **Settings → Monitor** and click **"Generate Agent Token"**. Copy the token — you'll only see it once.

### 2. Run the agent

The fastest way is Docker:

```bash
docker run -d \
  --name babything-agent \
  --restart unless-stopped \
  -e CLOUD_URL=wss://yourfamily.babything.app/monitor/agent \
  -e RTSP_URL=rtsp://192.168.1.50:554/stream1 \
  -e AGENT_TOKEN=your-token-here \
  ghcr.io/babything/agent:latest
```

Or run the binary directly:

```bash
export CLOUD_URL=wss://yourfamily.babything.app/monitor/agent
export RTSP_URL=rtsp://192.168.1.50:554/stream1
export AGENT_TOKEN=your-token-here
./babything-agent
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

## Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `CLOUD_URL` | **Yes** | WebSocket endpoint of your tenant | `wss://smith.babything.app/monitor/agent` |
| `RTSP_URL` | **Yes** | Full RTSP URL of your camera | `rtsp://admin:pass@192.168.1.50/stream` |
| `AGENT_TOKEN` | **Yes** | Agent token from admin dashboard | `eyJhbG...` |

---

## Building from Source

Requires **Go 1.23+** and **ffmpeg**.

```bash
cd babything-agent
go mod tidy
go build -o babything-agent .
```

### Cross-compilation

**Raspberry Pi (64-bit):**
```bash
GOOS=linux GOARCH=arm64 go build -o babything-agent-arm64 .
```

**Raspberry Pi (32-bit):**
```bash
GOOS=linux GOARCH=arm GOARM=7 go build -o babything-agent-arm .
```

**Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -o babything-agent.exe .
```

### Docker build

```bash
docker build -t babything-agent .
```

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| "Agent disconnected" in Monitor tab | Agent can't reach cloud | Check `CLOUD_URL` and outbound HTTPS/WSS |
| "Connection timed out" | WebRTC signaling failed | Check that agent is running and token is valid |
| Black screen, no video | Camera codec not H.264 | Verify camera outputs H.264, or ffmpeg will transcode |
| Choppy / high CPU | ffmpeg transcoding | Lower camera resolution or ensure `-c:v copy` works |
| Works on WiFi, not on cellular | NAT / firewall blocking P2P | Enable TURN server in cloud deployment |

### Check agent logs

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
