package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"github.com/pion/webrtc/v4"
)

// bindUDP binds a random local UDP port and returns the connection + port.
func bindUDP() (*net.UDPConn, int, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return nil, 0, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, 0, err
	}
	// Increase socket buffer so a high-bitrate 4K stream doesn't overflow
	// the default ~208KB Linux UDP buffer between reads.
	if err := conn.SetReadBuffer(2 * 1024 * 1024); err != nil {
		log.Printf("warning: failed to set udp read buffer: %v", err)
	}
	port := conn.LocalAddr().(*net.UDPAddr).Port
	return conn, port, nil
}

// startMediaRelay runs ffmpeg to pull the RTSP stream and feeds both video
// and audio into the WebRTC tracks.  Video is copied as-is (H.264 required);
// audio is transcoded to Opus for universal browser compatibility.
// If the camera has no audio stream ffmpeg continues with video only.
func startMediaRelay(ctx context.Context, rtspURL string, videoTrack, audioTrack *webrtc.TrackLocalStaticRTP) {
	relayCtx, cancelRelay := context.WithCancel(ctx)
	defer cancelRelay()

	videoConn, videoPort, err := bindUDP()
	if err != nil {
		log.Printf("video udp bind failed: %v", err)
		return
	}

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-re",
		"-i", rtspURL,
		"-map", "0:v:0",
		"-c:v", "copy",
		"-an",
		"-f", "rtp",
		fmt.Sprintf("rtp://127.0.0.1:%d?pkt_size=1200", videoPort),
	}

	var audioConn *net.UDPConn
	if audioTrack != nil {
		var audioPort int
		audioConn, audioPort, err = bindUDP()
		if err != nil {
			videoConn.Close()
			log.Printf("audio udp bind failed: %v", err)
			return
		}
		args = append(args,
			"-map", "0:a:0?",
			"-c:a", "libopus",
			"-application", "audio",
			"-b:a", "64k",
			"-vbr", "on",
			"-frame_duration", "20",
			"-f", "rtp",
			fmt.Sprintf("rtp://127.0.0.1:%d?pkt_size=1200", audioPort),
		)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = log.Writer()

	if err := cmd.Start(); err != nil {
		videoConn.Close()
		if audioConn != nil {
			audioConn.Close()
		}
		log.Printf("ffmpeg start failed: %v", err)
		return
	}
	log.Printf("ffmpeg started, relaying camera %s", rtspURL)

	// Monitor ffmpeg exit in a background goroutine.
	ffmpegDone := make(chan error, 1)
	go func() {
		ffmpegDone <- cmd.Wait()
	}()

	// Ensure ffmpeg is killed when the relay stops.
	defer func() {
		_ = cmd.Process.Kill()
	}()

	// Relay video RTP in the background.
	go relayRTP(relayCtx, videoConn, videoTrack, "video")

	// Relay audio RTP in the background.
	if audioConn != nil && audioTrack != nil {
		go relayRTP(relayCtx, audioConn, audioTrack, "audio")
	}

	// Block until ffmpeg exits.
	err = <-ffmpegDone
	if err != nil {
		log.Printf("ffmpeg exited with error: %v", err)
	} else {
		log.Printf("ffmpeg exited cleanly")
	}
}

// relayRTP reads RTP packets from a UDP socket and writes them to a WebRTC track.
func relayRTP(ctx context.Context, conn *net.UDPConn, track *webrtc.TrackLocalStaticRTP, label string) {
	defer conn.Close()
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
			log.Printf("%s udp read error: %v", label, err)
			return
		}

		if _, err := track.Write(buf[:n]); err != nil {
			log.Printf("%s track write error: %v", label, err)
			continue
		}
	}
}
