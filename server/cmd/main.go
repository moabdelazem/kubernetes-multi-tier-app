package main

import (
	"log"
	"net/http"

	"github.com/moabdelazem/k8s-app/internal/api"
	"github.com/moabdelazem/k8s-app/internal/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	router := api.SetupRoutes()
	log.Printf("Start on %v Environment: %s", cfg.Addr, cfg.Env)
	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		log.Fatalf("Something wrong happend %v", err)
	}
}
