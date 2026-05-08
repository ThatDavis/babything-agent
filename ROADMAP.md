# Babything Agent Roadmap

> Long-term direction for the babything-agent project.
>
> For the current session's work items, see `PLAN.md`.

---

## Now (Next 1–2 Weeks)

### Multi-Viewer Support ✅ In Progress
- [x] Share one video track across multiple peer connections
- [x] Single ffmpeg process regardless of viewer count
- [x] Per-viewer peer connection with auto-cleanup on disconnect
- [x] Thread-safe peer map with deadlock-free close paths

### CI/CD & Release Automation
- [ ] GitHub Actions workflow: build release binaries on tag push
  - [ ] Linux AMD64
  - [ ] Linux ARM64 (Raspberry Pi 4/5)
  - [ ] Linux ARMv7 (Raspberry Pi 3/Zero 2 W)
  - [ ] Windows AMD64
  - [ ] macOS AMD64 + ARM64 (Apple Silicon)
- [ ] GitHub Actions workflow: build and push Docker image to GHCR
  - [ ] Multi-arch image (amd64, arm64, arm/v7)
  - [ ] Tagged releases (`:latest`, `:v1.2.3`)
- [ ] Release notes automation from conventional commits

### Audio Stream Support ✅
- [x] Add Opus audio track to WebRTC peer connection
- [x] FFmpeg args: dual-output H.264 video + Opus audio from RTSP
- [x] Browser-side audio playback with audio-only mode toggle
- [x] Screen wake lock while monitoring
- [x] Payload-type rewrite to match Pion default MediaEngine

---

## Next (1–3 Months)

### Multi-Camera Support
- [ ] Config file (`config.yaml` or `config.json`) instead of env vars only
- [ ] Support multiple camera streams per agent
- [ ] Browser UI: camera selector in Monitor tab
- [ ] Independent WebRTC peer connections per camera

### Camera Auto-Discovery
- [ ] ONVIF discovery: scan local network for ONVIF-compatible cameras
- [ ] mDNS/Bonjour discovery for common consumer cameras
- [ ] Print discovered cameras with RTSP URLs to logs

### Motion Detection & Alerts
- [ ] Optional motion detection via FFmpeg `select=gt(scene\,0.003)` or lightweight frame diff
- [ ] Push alert to cloud API when motion detected
- [ ] Cloud stores short clip / snapshot and sends push notification to caregivers
- [ ] Configurable motion sensitivity and quiet hours

---

## Later (3–6 Months)

### Reliability & Observability
- [ ] Prometheus metrics endpoint (`/metrics`)
  - Connected status, bitrate, frame rate, reconnect count
- [ ] Health check endpoint for Docker / orchestrators
- [ ] Structured JSON logging with configurable level
- [ ] Graceful degradation: MJPEG fallback when WebRTC fails

### Security Hardening
- [ ] Token refresh: rotate agent token without restarting agent
- [ ] mTLS option for agent-to-cloud WebSocket (enterprise/self-hosted)
- [ ] IP allowlist: restrict which IPs can connect as agents

### Platform Packaging ✅
- [x] Native binaries for Linux, Windows, macOS
- [x] systemd service file for Linux bare-metal installs
- [x] Wails desktop app for Windows and macOS
- [x] GitHub Actions release pipeline with build matrix
- [ ] Homebrew formula for macOS — deferred
- [ ] Windows service wrapper — deferred
- [ ] Raspberry Pi Imager-compatible image (agent pre-installed) — deferred

---

## Non-Goals

- Cloud recording / DVR (out of scope for the agent; belongs in cloud API)
- AI person/baby detection (too heavy for lightweight agent)
- Two-way audio (intercom) — deferred until requested

---

*Last updated: 2026-05-02*
