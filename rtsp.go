package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/pion/webrtc/v4"
)

// processAlive checks whether a process is still running (Unix only).
func processAlive(p *os.Process) bool {
	err := p.Signal(syscall.Signal(0))
	return err == nil
}

// startRTPRelay runs ffmpeg to pull the RTSP stream and copy the raw RTP
// packets straight into the WebRTC track.  It requires the camera to emit
// H.264; if your camera uses a different codec ffmpeg will exit and the caller
// is expected to restart the relay.
func startRTPRelay(ctx context.Context, rtspURL string, track *webrtc.TrackLocalStaticRTP) {
	// Bind a random local UDP port for ffmpeg to send RTP to.
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		log.Printf("resolve udp addr failed: %v", err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("listen udp failed: %v", err)
		return
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	rtpURL := fmt.Sprintf("rtp://127.0.0.1:%d?pkt_size=1200", localAddr.Port)

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-re",
		"-i", rtspURL,
		"-c:v", "copy",
		"-an",
		"-f", "rtp",
		rtpURL,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = log.Writer()

	if err := cmd.Start(); err != nil {
		log.Printf("ffmpeg start failed: %v", err)
		return
	}

	// Ensure ffmpeg is killed when the relay stops.
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	buf := make([]byte, 1500)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !processAlive(cmd.Process) {
			log.Printf("ffmpeg process died, stopping relay")
			return
		}

		_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("udp read error: %v", err)
			return
		}

		if _, err := track.Write(buf[:n]); err != nil {
			log.Printf("track write error: %v", err)
			return
		}
	}
}
