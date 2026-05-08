package agent

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
)

// startMediaRelay runs ffmpeg to pull the RTSP stream and copy the raw RTP
// packets into the WebRTC tracks.  It outputs both H.264 video and Opus audio
// to separate local UDP ports so the two streams never interleave.
func startMediaRelay(ctx context.Context, rtspURL string, videoTrack, audioTrack *webrtc.TrackLocalStaticRTP) {
	// Bind two random local UDP ports — one for video, one for audio.
	videoConn, videoPort, err := bindUDP("127.0.0.1:0")
	if err != nil {
		log.Printf("bind video udp failed: %v", err)
		return
	}
	defer videoConn.Close()

	audioConn, audioPort, err := bindUDP("127.0.0.1:0")
	if err != nil {
		log.Printf("bind audio udp failed: %v", err)
		return
	}
	defer audioConn.Close()

	videoRTP := fmt.Sprintf("rtp://127.0.0.1:%d?pkt_size=1200", videoPort)
	audioRTP := fmt.Sprintf("rtp://127.0.0.1:%d?pkt_size=1200", audioPort)

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-re",
		"-i", rtspURL,
		// Video output: copy H.264, no audio
		"-c:v", "copy",
		"-an",
		"-f", "rtp",
		videoRTP,
		// Audio output: transcode to Opus, no video
		"-c:a", "libopus",
		"-ar", "48000",
		"-ac", "2",
		"-vn",
		"-f", "rtp",
		audioRTP,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = log.Writer()

	if err := cmd.Start(); err != nil {
		log.Printf("ffmpeg start failed: %v", err)
		return
	}
	log.Printf("ffmpeg started, video→udp:%d audio→udp:%d", videoPort, audioPort)

	// Monitor ffmpeg exit in a background goroutine.
	ffmpegDone := make(chan error, 1)
	go func() {
		ffmpegDone <- cmd.Wait()
	}()

	// Ensure ffmpeg is killed when the relay stops.
	defer func() {
		_ = cmd.Process.Kill()
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	// Video relay — ffmpeg assigns PT 96 for the first output, which matches
	// Pion's default H.264 payload type. No rewrite needed.
	go func() {
		defer wg.Done()
		relayRTP(ctx, videoConn, videoTrack, 0)
	}()

	// Audio relay — ffmpeg assigns PT 97 for the second output, but Pion's
	// default MediaEngine expects Opus at PT 111. Rewrite the PT byte.
	go func() {
		defer wg.Done()
		relayRTP(ctx, audioConn, audioTrack, 111)
	}()

	// Wait for either context cancellation or ffmpeg exit.
	select {
	case <-ctx.Done():
		return
	case err := <-ffmpegDone:
		if err != nil {
			log.Printf("ffmpeg exited with error: %v", err)
		} else {
			log.Printf("ffmpeg exited cleanly")
		}
	}

	// Closing the sockets unblocks the relay goroutines so they can exit.
	videoConn.Close()
	audioConn.Close()
	wg.Wait()
}

// bindUDP creates a UDP listener on the given address and returns the
// connection plus the assigned port.
func bindUDP(addr string) (*net.UDPConn, int, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, 0, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, 0, err
	}
	if err := conn.SetReadBuffer(2 * 1024 * 1024); err != nil {
		log.Printf("warning: failed to set udp read buffer: %v", err)
	}
	port := conn.LocalAddr().(*net.UDPAddr).Port
	return conn, port, nil
}

// rewritePayloadType changes the RTP payload type in raw packet bytes.
// It preserves the marker bit (top bit of the second byte).
func rewritePayloadType(pkt []byte, newPT uint8) {
	if len(pkt) < 2 {
		return
	}
	pkt[1] = (pkt[1] & 0x80) | (newPT & 0x7F)
}

// relayRTP reads RTP packets from conn and writes them to track.
// If rewritePT > 0, the payload type in each packet is rewritten to that value.
func relayRTP(ctx context.Context, conn *net.UDPConn, track *webrtc.TrackLocalStaticRTP, rewritePT uint8) {
	buf := make([]byte, 1500)
	for {
		select {
		case <-ctx.Done():
			return
		default:
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

		if rewritePT > 0 {
			rewritePayloadType(buf[:n], rewritePT)
		}

		if _, err := track.Write(buf[:n]); err != nil {
			log.Printf("track write error: %v", err)
			return
		}
	}
}
