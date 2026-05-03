package main

import (
	"log"
	"os"
	"os/signal"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	log.Printf("babything agent starting")
	log.Printf("cloud: %s", cfg.CloudURL)
	log.Printf("camera: %s", cfg.RTSPURL)

	client := NewSignalingClient(cfg)
	go client.Run()

	// Wait for interrupt
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	log.Println("shutting down...")
	client.Stop()
	log.Println("goodbye")
}
