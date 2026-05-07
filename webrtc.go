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
// cloudServers are sent automatically by the cloud over the signalling socket.
// If the cloud has not sent any config yet, the optional env-var TURN settings
// are used as a fallback.
// onCandidate is called for each locally gathered ICE candidate so they can be
// sent to the peer via trickle ICE.
func NewPeerConnection(cloudServers []webrtc.ICEServer, turnURL, turnUser, turnPass string, onCandidate func(*webrtc.ICECandidate)) (*PeerConnection, error) {
	ctx, cancel := context.WithCancel(context.Background())

	var iceServers []webrtc.ICEServer
	if len(cloudServers) > 0 {
		iceServers = append([]webrtc.ICEServer{}, cloudServers...)
	} else {
		iceServers = []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		}
		if turnURL != "" {
			iceServers = append(iceServers, webrtc.ICEServer{
				URLs:       []string{turnURL},
				Username:   turnUser,
				Credential: turnPass,
			})
		}
	}

	config := webrtc.Configuration{
		ICEServers: iceServers,
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		cancel()
		return nil, err
	}

	if onCandidate != nil {
		pc.OnICECandidate(onCandidate)
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

// CreateAnswer generates an SDP answer and begins ICE gathering.
// Candidates are sent asynchronously via the OnICECandidate callback
// (trickle ICE), so the answer is returned immediately.
func (p *PeerConnection) CreateAnswer() (string, error) {
	answer, err := p.pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	if err := p.pc.SetLocalDescription(answer); err != nil {
		return "", err
	}

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
// If ffmpeg exits (e.g. camera drops) the relay is automatically restarted.
func (p *PeerConnection) StartRTSP(rtspURL string) {
	go func() {
		for {
			select {
			case <-p.ctx.Done():
				return
			default:
			}
			startRTPRelay(p.ctx, rtspURL, p.videoTrack)
			// ffmpeg exited; wait a moment before restarting to avoid busy-looping
			select {
			case <-p.ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}()
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
