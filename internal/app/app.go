package app

import (
	"log"
	"loyalty-service/internal/config"
)

func Run(args []string) {

	cfg, err := config.New(args)
	if err != nil {
		log.Fatalf("failed to create app config: %s", err)
	}

	if cfg == nil {
		log.Fatalf("app config is nil")
	}
	log.Printf("Server started on %s", cfg.RunAddress)
}
