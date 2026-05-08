# Babything Agent — Session Plan

> Current work items for this session.
> For the long-term roadmap, see `ROADMAP.md`.

---

## Goal

Refactor the babything-agent to support multiple simultaneous viewers with a shared video track and single ffmpeg process, laying the groundwork for audio.

---

## This Session

- [x] **Multi-viewer shared track refactor** ✅ Done
  - [x] Create `MediaSource` to own shared tracks and ffmpeg lifecycle
  - [x] Refactor `PeerConnection` to accept shared tracks instead of creating its own
  - [x] Replace single `pc` with `peers` map keyed by `watchID`
  - [x] Route ICE candidates to the correct peer connection
  - [x] Auto-cleanup dead peers via `OnConnectionStateChange`
  - [x] Start ffmpeg once on first offer, share across all viewers
  - [x] Avoid mutex deadlocks during peer close / reconnect

- [ ] **GitHub Actions: release binaries**
  - [ ] Create `.github/workflows/release.yml`
  - [ ] Build matrix: linux/amd64, linux/arm64, linux/arm, windows/amd64, darwin/amd64, darwin/arm64
  - [ ] Attach binaries to GitHub Release on tag push
  - [ ] Test workflow with a dummy tag

- [ ] **GitHub Actions: Docker image**
  - [ ] Create `.github/workflows/docker.yml`
  - [ ] Multi-arch build (amd64, arm64, arm/v7) via buildx
  - [ ] Push to GHCR on every push to `main` and on tags

- [x] **Audio stream support**
  - [x] Add audio track to `media.go`
  - [x] Update `rtsp.go` ffmpeg args to extract audio (transcode to opus)
  - [x] Browser receives and plays audio (transceiver added in MonitorTab)
  - [ ] Update README with audio notes

---

## Done This Session

*(Fill in as items complete)*

---

## Blockers / Notes

- FFmpeg must be available in the GitHub Actions runner (use `ubuntu-latest` which has it pre-installed)
- For cross-compilation, Go cross-compiles easily but Docker buildx is needed for multi-arch images
- Audio codec compatibility: most cameras use AAC, which browsers support via WebRTC. If not, re-encode to Opus.

---

*Created: 2026-05-02*
