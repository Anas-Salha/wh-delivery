package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/anas-salha/wh-delivery/delivery/internal/httpapi"
	"github.com/anas-salha/wh-delivery/delivery/internal/repo"
	"github.com/anas-salha/wh-delivery/delivery/internal/service"

	"github.com/anas-salha/wh-delivery/delivery/internal/config"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	store := repo.NewPostgres(db)
	webhooksSvc := service.NewWebhooksService(store)
	sourcesSvc := service.NewSourcesService(store, store)

	services := httpapi.Services{
		Webhooks: webhooksSvc,
		Sources:  sourcesSvc,
	}

	if err := httpapi.Run(cfg, services); err != nil {
		log.Fatalf("[%s] server error: %v", cfg.ServiceName, err)
	}
}
