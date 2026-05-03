package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/pion/webrtc/v4"
)

// PeerConnection wraps a Pion peer connection for the monitor agent.
type PeerConnection struct {
	pc         *webrtc.PeerConnection
	videoTrack *webrtc.TrackLocalStaticRTP
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewPeerConnection creates a new WebRTC peer connection.
func NewPeerConnection() (*PeerConnection, error) {
	ctx, cancel := context.WithCancel(context.Background())

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		cancel()
		return nil, err
	}

	// Create a video track (H.264)
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
		"video",
		"babything-monitor",
	)
	if err != nil {
		pc.Close()
		cancel()
		return nil, err
	}

	if _, err := pc.AddTrack(videoTrack); err != nil {
		pc.Close()
		cancel()
		return nil, err
	}

	return &PeerConnection{
		pc:         pc,
		videoTrack: videoTrack,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// SetRemoteDescription accepts the browser's SDP offer.
func (p *PeerConnection) SetRemoteDescription(sdp string) error {
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp,
	}
	return p.pc.SetRemoteDescription(offer)
}

// CreateAnswer generates an SDP answer and waits for ICE gathering to finish.
func (p *PeerConnection) CreateAnswer() (string, error) {
	answer, err := p.pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	if err := p.pc.SetLocalDescription(answer); err != nil {
		return "", err
	}

	// Block until ICE gathering is complete
	<-webrtc.GatheringCompletePromise(p.pc)

	return p.pc.LocalDescription().SDP, nil
}

// AddICECandidate injects a remote ICE candidate.
func (p *PeerConnection) AddICECandidate(candidate json.RawMessage) error {
	var c webrtc.ICECandidateInit
	if err := json.Unmarshal(candidate, &c); err != nil {
		return err
	}
	return p.pc.AddICECandidate(c)
}

// StartRTSP launches ffmpeg to read the camera and forward RTP into the track.
func (p *PeerConnection) StartRTSP(rtspURL string) {
	go startRTPRelay(p.ctx, rtspURL, p.videoTrack)
}

// Close tears down the peer connection and stops ffmpeg.
func (p *PeerConnection) Close() {
	p.cancel()
	if p.pc != nil {
		if err := p.pc.Close(); err != nil {
			log.Printf("pc close error: %v", err)
		}
	}
}
