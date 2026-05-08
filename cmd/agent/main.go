package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/babything/agent/internal/agent"
)

func main() {
	cfg, err := agent.LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	log.Printf("babything agent starting")
	log.Printf("cloud: %s", cfg.CloudURL)
	log.Printf("camera: %s", cfg.RTSPURL)

	client, err := agent.NewSignalingClient(cfg)
	if err != nil {
		log.Fatalf("failed to create signaling client: %v", err)
	}
	go client.Run()

	// Wait for interrupt
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	log.Println("shutting down...")
	client.Stop()
	log.Println("goodbye")
}
