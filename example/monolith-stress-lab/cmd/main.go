package main

import (
	"log"

	"github.com/codenbase/kubernetes/example/monolith-stress-lab/cmd/app"
)

func main() {
	cfg := app.LoadConfig()

	db := app.InitDB(cfg.DSN)

	r := app.SetupRouter(db, cfg.JWTSecret)
	
	log.Printf("Starting Gin web service on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
