# Plan: Native Agent (Linux / Windows / macOS)

> Pivot from Docker-only to native executables for all three platforms.
> End users download and run a binary — no Docker, no networking headaches.
>
> Branch: `feature/native-agent`

---

## Problem Statement

The Docker-based agent has recurring networking issues:
- Virtual interfaces (`tap0`, `docker0`, `veth*`) leak into WebRTC ICE candidates
- Host networking mode behaves inconsistently across environments
- End users find Docker intimidating

**Goal:** Ship native binaries that "just work" with zero container networking complexity.

---

## Proposed Architecture

### Package Refactor

Current: everything lives in `package main`.

New layout:

```
babything-agent/
├── cmd/
│   ├── agent/           # headless binary (all platforms)
│   │   └── main.go      # loads config, starts agent, blocks on signal
│   └── desktop/         # desktop wrapper (Windows + macOS)
│       └── main.go      # Wails app: tray/menu bar + settings window
├── internal/
│   ├── agent/           # core logic (extracted from current main package)
│   │   ├── client.go    # SignalingClient
│   │   ├── config.go    # Config struct + loading
│   │   ├── media.go     # MediaSource
│   │   ├── peer.go      # PeerConnection wrapper
│   │   ├── rtsp.go      # ffmpeg RTSP relay
│   │   └── signaling.go # WebSocket signaling
│   └── configfile/      # YAML/JSON config file loader
├── frontend/            # Wails web frontend (settings UI)
│   ├── src/
│   └── wails.json
├── build/
│   ├── linux/
│   │   └── babything-agent.service
│   ├── windows/
│   │   └── installer.nsi   # optional NSIS installer
│   └── darwin/
│       └── Info.plist
├── plan-native.md       # this file
├── go.mod
├── go.sum
└── wails.json
```

### Core Agent (`internal/agent/`)

Extract all non-CLI logic from the current root `.go` files into an `internal/agent` package:

- `Config` struct supports **both** env vars and config file (config file wins)
- `SignalingClient` exposes `Start() / Stop() / Status()` for the GUI to call
- No changes to WebRTC, RTSP, or signaling logic

### Desktop App (`cmd/desktop/` + `frontend/`)

**Framework: Wails v2**

Why Wails:
- Go backend — reuses `internal/agent` directly, no separate process
- Web frontend — easy settings form with HTML/CSS (or React/Vue if desired)
- System tray on Windows, menu bar on macOS — built-in support
- Native window chrome — not Electron-heavy
- Cross-compiles via Go build tags

**UX:**
- App starts hidden (no window on launch)
- Tray/menu bar icon shows status: 🟢 running / 🔴 stopped / ⚠️ error
- Left-click (or single menu item): "Open Settings"
- Settings window: small form with Cloud URL, RTSP URL, Agent Token, TURN optional
- "Start / Stop" button
- "Quit" exits completely
- Config auto-saves to `~/.config/babything/agent.yaml` (or platform equivalent)

### Headless Binary (`cmd/agent/`)

- Linux primary delivery method (systemd service)
- Also usable on Windows/macOS for power users
- Reads config file, falls back to env vars
- Logs to stdout/stderr (captured by systemd journal)

---

## Platform Details

### Linux

- **Binary:** `babything-agent` (amd64, arm64, arm/v7)
- **Config:** `~/.config/babything/agent.yaml` or `/etc/babything/agent.yaml`
- **Service:** systemd unit file (`babything-agent.service`)
  - Auto-start on boot
  - Restart on failure
  - `journalctl -u babything-agent` for logs
- **No GUI** — config file + CLI flags only
- **Install:** download tarball, extract, `sudo systemctl enable --now babything-agent`

### Windows

- **Binary:** `BabythingAgent.exe`
- **Config:** `%APPDATA%\Babything\agent.yaml`
- **GUI:** System tray icon in notification area
  - Right-click menu: Settings | Start | Stop | Quit
  - Settings window: form fields + save button
  - Auto-starts agent on login (optional checkbox)
- **Install:** single `.exe` (portable) or MSI installer

### macOS

- **Binary:** `BabythingAgent.app` (bundled)
- **Config:** `~/Library/Application Support/Babything/agent.yaml`
- **GUI:** Menu bar icon (top-right)
  - Dropdown menu: Settings | Start | Stop | Quit
  - Settings window: same form as Windows
  - Auto-start on login via LaunchAgent
- **Install:** `.app` bundle or `.dmg`

---

## Config System

Replace env-var-only with **config file primary, env vars fallback**.

Example `agent.yaml`:

```yaml
cloud_url: "wss://smith.babything.app/monitor/agent"
rtsp_url: "rtsp://admin:pass@192.168.1.50:554/stream1"
agent_token: "eyJhbG..."

# Optional
turn_url: ""
turn_username: ""
turn_password: ""

# Advanced
log_level: info
```

Env vars still work for backward compatibility and Docker (if kept):
- `CLOUD_URL` → `cloud_url`
- `RTSP_URL` → `rtsp_url`
- etc.

---

## Build & Release

### GitHub Actions Matrix

Extend the existing (incomplete) release workflow:

| Target | OS | Arch | Output |
|--------|-----|------|--------|
| linux-amd64 | Linux | amd64 | binary + systemd file |
| linux-arm64 | Linux | arm64 | binary + systemd file |
| linux-arm | Linux | arm/v7 | binary + systemd file |
| windows-amd64 | Windows | amd64 | `.exe` (Wails build) |
| darwin-amd64 | macOS | amd64 | `.app` bundle |
| darwin-arm64 | macOS | arm64 | `.app` bundle (Apple Silicon) |

### Wails Build Requirements

- Windows runner: needs WebView2 runtime (pre-installed on `windows-latest`)
- macOS runner: needs Xcode (pre-installed on `macos-latest`)
- Linux runner: for headless binary only (no Wails GUI)

### Distribution

- GitHub Releases page with per-platform assets
- Optional: Homebrew formula for macOS
- Optional: Chocolatey/winget for Windows
- Optional: `.deb` / `.rpm` packages for Linux

---

## Phases

### Phase 1 — Refactor Core (foundation)
- [ ] Extract `internal/agent/` package from current root `.go` files
- [ ] Move `main.go` to `cmd/agent/main.go`
- [ ] Add `internal/configfile/` with YAML loader + env fallback
- [ ] Verify headless binary still works (Linux)

### Phase 2 — Desktop Scaffold (Win + Mac)
- [ ] Initialize Wails v2 project in `cmd/desktop/`
- [ ] Wire `internal/agent` into Wails runtime (Start/Stop/Status bindings)
- [ ] Build minimal settings form in `frontend/`
- [ ] System tray (Windows) and menu bar (macOS) integration
- [ ] Config read/write through GUI

### Phase 3 — Platform Packaging
- [ ] Linux: systemd service file + install script
- [ ] Windows: Wails build + optional NSIS installer
- [ ] macOS: Wails build + `.app` bundling + codesign (if cert available)

### Phase 4 — CI/CD Release Pipeline
- [ ] Update `.github/workflows/release.yml` with full build matrix
- [ ] Attach all platform binaries to GitHub Release on tag push
- [ ] Test end-to-end on each platform (manual QA)

### Phase 5 — Cleanup
- [ ] Update `README.md` with native install instructions
- [ ] Update `ROADMAP.md` — mark Platform Packaging complete
- [ ] Decide fate of Docker (keep as alternative vs. deprecate vs. remove)

---

## Open Decisions

Before implementation starts, confirm:

1. **GUI framework:** Wails v2 (recommended) vs. Fyne vs. native bindings?
2. **Docker fate:** Keep Dockerfile as an alternative, or remove entirely?
3. **Frontend stack:** Plain HTML/CSS/JS (lightweight) or small React/Vue app?
4. **Config format:** YAML (human-friendly) vs. JSON vs. TOML?

---

## Risks

| Risk | Mitigation |
|------|------------|
| Wails adds build complexity (WebView2, Xcode deps) | Use GitHub Actions runners with pre-installed deps; Linux stays headless |
| Refactoring breaks existing agent logic | Keep WebRTC/signaling code unchanged; only move files |
| macOS codesigning requires Apple Developer cert | Start with unsigned `.app`; add codesign later |
| Config file path varies across distros | Use XDG dirs on Linux, known Windows/macOS paths |

---

*Created: 2026-05-07*
