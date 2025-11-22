package main

import (
	"log"
	"github.com/PaulLocust/Avito-review/config"

)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run TODO
	// app.Run(cfg)
}
