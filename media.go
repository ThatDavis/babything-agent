package main

import (
	"context"
	"time"

	"github.com/pion/webrtc/v4"
)

// MediaSource holds the shared media tracks and manages the ffmpeg RTSP relay.
// One MediaSource exists per agent. Its tracks are added to every viewer's
// peer connection, so ffmpeg runs exactly once regardless of viewer count.
type MediaSource struct {
	videoTrack *webrtc.TrackLocalStaticRTP
	audioTrack *webrtc.TrackLocalStaticRTP
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewMediaSource creates the shared video and audio tracks.
func NewMediaSource() (*MediaSource, error) {
	ctx, cancel := context.WithCancel(context.Background())

	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
		"video",
		"babything-monitor-video",
	)
	if err != nil {
		cancel()
		return nil, err
	}

	audioTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio",
		"babything-monitor-audio",
	)
	if err != nil {
		cancel()
		return nil, err
	}

	return &MediaSource{
		videoTrack: videoTrack,
		audioTrack: audioTrack,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Tracks returns all active tracks to be added to peer connections.
func (m *MediaSource) Tracks() []*webrtc.TrackLocalStaticRTP {
	tracks := []*webrtc.TrackLocalStaticRTP{m.videoTrack}
	if m.audioTrack != nil {
		tracks = append(tracks, m.audioTrack)
	}
	return tracks
}

// StartRTSP launches ffmpeg and feeds RTP into the shared tracks.
// It restarts automatically if ffmpeg exits.
func (m *MediaSource) StartRTSP(rtspURL string) {
	go func() {
		for {
			select {
			case <-m.ctx.Done():
				return
			default:
			}
			startMediaRelay(m.ctx, rtspURL, m.videoTrack, m.audioTrack)
			// ffmpeg exited; wait before restarting to avoid busy-looping
			select {
			case <-m.ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}()
}

// Close stops the RTSP relay and releases resources.
func (m *MediaSource) Close() {
	m.cancel()
}
