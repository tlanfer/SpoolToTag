package main

import (
	"fmt"
	"os"
)

type Config struct {
	APIKey     string
	ListenAddr string
	Model      string
}

func LoadConfig() (Config, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return Config{}, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o"
	}

	return Config{
		APIKey:     apiKey,
		ListenAddr: listenAddr,
		Model:      model,
	}, nil
}
