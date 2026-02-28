package main

import (
	"log"
	"net/http"

	"github.com/tlanfer/SpoolToTag/handler"
	"github.com/tlanfer/SpoolToTag/openai"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	client := openai.NewClient(cfg.APIKey, cfg.Model)
	h := handler.New(client)

	log.Printf("listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, h); err != nil {
		log.Fatal(err)
	}
}
