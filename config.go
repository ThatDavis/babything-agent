package main

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration for the babything agent.
type Config struct {
	CloudURL      string // e.g. wss://smith.babything.app/monitor/agent
	RTSPURL       string // e.g. rtsp://192.168.1.50:554/stream1
	Token         string // JWT agent token from admin dashboard
	TURNURL       string // e.g. turn:turn.babything.app:3478
	TURNUsername  string // TURN server username
	TURNPassword  string // TURN server password
}

func loadConfig() (Config, error) {
	cfg := Config{
		CloudURL:     os.Getenv("CLOUD_URL"),
		RTSPURL:      os.Getenv("RTSP_URL"),
		Token:        os.Getenv("AGENT_TOKEN"),
		TURNURL:      os.Getenv("TURN_URL"),
		TURNUsername: os.Getenv("TURN_USERNAME"),
		TURNPassword: os.Getenv("TURN_PASSWORD"),
	}
	if cfg.CloudURL == "" {
		return cfg, fmt.Errorf("CLOUD_URL is required")
	}
	if cfg.RTSPURL == "" {
		return cfg, fmt.Errorf("RTSP_URL is required")
	}
	if cfg.Token == "" {
		return cfg, fmt.Errorf("AGENT_TOKEN is required")
	}
	return cfg, nil
}
