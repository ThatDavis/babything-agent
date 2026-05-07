package main

import (
	"encoding/json"
	"log"

	"github.com/pion/webrtc/v4"
)

// PeerConnection wraps a single Pion peer connection for one viewer.
type PeerConnection struct {
	pc *webrtc.PeerConnection
}

// NewPeerConnection creates a WebRTC peer connection and adds the given tracks.
func NewPeerConnection(cloudServers []webrtc.ICEServer, turnURL, turnUser, turnPass string, tracks []*webrtc.TrackLocalStaticRTP, onCandidate func(*webrtc.ICECandidate)) (*PeerConnection, error) {
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
		return nil, err
	}

	if onCandidate != nil {
		pc.OnICECandidate(onCandidate)
	}

	for _, track := range tracks {
		if _, err := pc.AddTrack(track); err != nil {
			pc.Close()
			return nil, err
		}
	}

	return &PeerConnection{pc: pc}, nil
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

// OnConnectionStateChange registers a callback for peer connection state changes.
func (p *PeerConnection) OnConnectionStateChange(f func(webrtc.PeerConnectionState)) {
	p.pc.OnConnectionStateChange(f)
}

// Close tears down the peer connection.
func (p *PeerConnection) Close() {
	if p.pc != nil {
		if err := p.pc.Close(); err != nil {
			log.Printf("pc close error: %v", err)
		}
	}
}
