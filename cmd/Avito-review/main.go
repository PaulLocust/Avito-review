package main

import (
	"log"

	"github.com/PaulLocust/Avito-review/config"
	"github.com/PaulLocust/Avito-review/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(cfg)
}
