package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

// SignalMessage is the JSON envelope exchanged with the cloud signal server.
type SignalMessage struct {
	Type      string          `json:"type"`
	WatchID   string          `json:"watchId,omitempty"`
	SDP       string          `json:"sdp,omitempty"`
	Candidate json.RawMessage `json:"candidate,omitempty"`
	SessionID string          `json:"sessionId,omitempty"`
}

// SignalingClient maintains the WebSocket connection to the cloud.
type SignalingClient struct {
	cfg        Config
	pc         *PeerConnection
	iceServers []webrtc.ICEServer

	conn    *websocket.Conn
	done    chan struct{}
	writeMu chan struct{} // serialises writes (capacity 1)
}

// NewSignalingClient creates a client that will keep reconnecting until Stop().
func NewSignalingClient(cfg Config) *SignalingClient {
	return &SignalingClient{
		cfg:     cfg,
		done:    make(chan struct{}),
		writeMu: make(chan struct{}, 1),
	}
}

// Run blocks, keeping a persistent connection to the cloud.
func (c *SignalingClient) Run() {
	backoff := 1 * time.Second
	for {
		select {
		case <-c.done:
			return
		default:
		}

		if err := c.connect(); err != nil {
			log.Printf("connect failed: %v, retrying in %v", err, backoff)
			select {
			case <-time.After(backoff):
			case <-c.done:
				return
			}
			if backoff < 30*time.Second {
				backoff += 2 * time.Second
			}
			continue
		}
		backoff = 1 * time.Second
		c.readLoop()

		// Clean up peer connection on disconnect so next offer starts fresh
		if c.pc != nil {
			c.pc.Close()
			c.pc = nil
		}
	}
}

func (c *SignalingClient) connect() error {
	u, err := url.Parse(c.cfg.CloudURL)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("token", c.cfg.Token)
	u.RawQuery = q.Encode()

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.Dial(u.String(), http.Header{})
	if err != nil {
		return err
	}
	c.conn = conn

	// Reset read deadline on native WebSocket ping frames so the connection
	// stays alive even when the server switches from JSON ping to native ping.
	c.conn.SetPingHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return c.conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
	})

	return nil
}

func (c *SignalingClient) readLoop() {
	for {
		select {
		case <-c.done:
			return
		default:
		}

		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("read error: %v", err)
			return
		}

		var msg SignalMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "offer":
			go c.handleOffer(msg)
		case "ice":
			go c.handleRemoteICE(msg)
		case "ping":
			c.send(SignalMessage{Type: "pong"})
		case "config":
			c.handleConfig(data)
		case "connected":
			log.Printf("connected to cloud, session %s", msg.SessionID)
		}
	}
}

func (c *SignalingClient) handleConfig(data []byte) {
	var payload struct {
		ICEServers []rawICEServer `json:"iceServers"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("failed to parse ice config: %v", err)
		return
	}

	servers := make([]webrtc.ICEServer, 0, len(payload.ICEServers))
	for _, s := range payload.ICEServers {
		urls, err := parseURLs(s.URLs)
		if err != nil {
			continue
		}
		servers = append(servers, webrtc.ICEServer{
			URLs:       urls,
			Username:   s.Username,
			Credential: s.Credential,
		})
	}

	if len(servers) > 0 {
		c.iceServers = servers
		log.Printf("received ice config with %d server(s)", len(servers))
	}
}

type rawICEServer struct {
	URLs       json.RawMessage `json:"urls"`
	Username   string          `json:"username,omitempty"`
	Credential string          `json:"credential,omitempty"`
}

func parseURLs(raw json.RawMessage) ([]string, error) {
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var single string
	if err := json.Unmarshal(raw, &single); err != nil {
		return nil, err
	}
	return []string{single}, nil
}

func (c *SignalingClient) handleOffer(msg SignalMessage) {
	log.Printf("received offer for watch %s", msg.WatchID)

	onCandidate := func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		init := candidate.ToJSON()
		raw, err := json.Marshal(init)
		if err != nil {
			log.Printf("marshal candidate failed: %v", err)
			return
		}
		c.send(SignalMessage{Type: "ice", WatchID: msg.WatchID, Candidate: raw})
	}

	pc, err := NewPeerConnection(c.iceServers, c.cfg.TURNURL, c.cfg.TURNUsername, c.cfg.TURNPassword, onCandidate)
	if err != nil {
		log.Printf("failed to create peer connection: %v", err)
		return
	}

	if c.pc != nil {
		c.pc.Close()
	}
	c.pc = pc

	if err := pc.SetRemoteDescription(msg.SDP); err != nil {
		log.Printf("set remote description failed: %v", err)
		return
	}

	answer, err := pc.CreateAnswer()
	if err != nil {
		log.Printf("create answer failed: %v", err)
		return
	}

	c.send(SignalMessage{Type: "answer", WatchID: msg.WatchID, SDP: answer})

	// Start pulling RTSP and feeding into the peer connection
	pc.StartRTSP(c.cfg.RTSPURL)
}

func (c *SignalingClient) handleRemoteICE(msg SignalMessage) {
	if c.pc == nil {
		return
	}
	if err := c.pc.AddICECandidate(msg.Candidate); err != nil {
		log.Printf("add ice candidate failed: %v", err)
	}
}

func (c *SignalingClient) send(msg SignalMessage) {
	select {
	case c.writeMu <- struct{}{}:
		if c.conn != nil {
			_ = c.conn.WriteJSON(msg)
		}
		<-c.writeMu
	default:
	}
}

// Stop terminates the client and closes all connections.
func (c *SignalingClient) Stop() {
	close(c.done)
	if c.conn != nil {
		c.conn.Close()
	}
	if c.pc != nil {
		c.pc.Close()
	}
}
