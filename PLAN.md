# Babything Agent — Session Plan

> Current work items for this session.
> For the long-term roadmap, see `ROADMAP.md`.

---

## Goal

Ship CI/CD release automation and audio stream support for the babything-agent.

---

## This Session

- [ ] **GitHub Actions: release binaries**
  - [ ] Create `.github/workflows/release.yml`
  - [ ] Build matrix: linux/amd64, linux/arm64, linux/arm, windows/amd64, darwin/amd64, darwin/arm64
  - [ ] Attach binaries to GitHub Release on tag push
  - [ ] Test workflow with a dummy tag

- [ ] **GitHub Actions: Docker image**
  - [ ] Create `.github/workflows/docker.yml`
  - [ ] Multi-arch build (amd64, arm64, arm/v7) via buildx
  - [ ] Push to GHCR on every push to `main` and on tags

- [ ] **Audio stream support**
  - [ ] Add audio track to `webrtc.go`
  - [ ] Update `rtsp.go` ffmpeg args to extract audio (copy or re-encode to opus)
  - [ ] Verify browser receives and plays audio
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
