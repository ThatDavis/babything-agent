package agent

import (
	"fmt"
	"os"

	"github.com/babything/agent/internal/configfile"
)

// Config holds all runtime configuration for the babything agent.
type Config struct {
	CloudURL      string `yaml:"cloud_url" json:"cloud_url"`
	RTSPURL       string `yaml:"rtsp_url" json:"rtsp_url"`
	Token         string `yaml:"agent_token" json:"agent_token"`
	TURNURL       string `yaml:"turn_url,omitempty" json:"turn_url,omitempty"`
	TURNUsername  string `yaml:"turn_username,omitempty" json:"turn_username,omitempty"`
	TURNPassword  string `yaml:"turn_password,omitempty" json:"turn_password,omitempty"`
	LogLevel      string `yaml:"log_level,omitempty" json:"log_level,omitempty"`
}

// LoadConfig loads configuration from environment variables and optionally
// from a YAML config file. Config file values override env vars.
func LoadConfig() (Config, error) {
	cfg := loadConfigFromEnv()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = configfile.DefaultPath()
	}

	if _, err := os.Stat(configPath); err == nil {
		if err := configfile.Load(configPath, &cfg); err != nil {
			return Config{}, fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}
	}

	if cfg.CloudURL == "" {
		return cfg, fmt.Errorf("cloud_url is required (set in config file or CLOUD_URL env var)")
	}
	if cfg.RTSPURL == "" {
		return cfg, fmt.Errorf("rtsp_url is required (set in config file or RTSP_URL env var)")
	}
	if cfg.Token == "" {
		return cfg, fmt.Errorf("agent_token is required (set in config file or AGENT_TOKEN env var)")
	}

	return cfg, nil
}

func loadConfigFromEnv() Config {
	turnPassword := os.Getenv("TURN_PASSWORD")
	if turnPassword == "" {
		turnPassword = os.Getenv("TURN_CREDENTIAL")
	}
	return Config{
		CloudURL:     os.Getenv("CLOUD_URL"),
		RTSPURL:      os.Getenv("RTSP_URL"),
		Token:        os.Getenv("AGENT_TOKEN"),
		TURNURL:      os.Getenv("TURN_URL"),
		TURNUsername: os.Getenv("TURN_USERNAME"),
		TURNPassword: turnPassword,
	}
}
