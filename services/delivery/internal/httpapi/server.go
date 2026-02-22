package httpapi

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/anas-salha/wh-delivery/delivery/internal/config"
)

type Server struct {
	serviceName string
}

type Services struct {
	Webhooks WebhooksService
	Sources  SourcesService
}

func Run(cfg config.Config, services Services) error {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf(
			"[%s] %s %s %d %s\n",
			cfg.ServiceName,
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	}))

	api := router.Group("/api/v1")
	NewWebhooksHandler(cfg.ServiceName, services.Webhooks).Register(api)
	NewSourcesHandler(cfg.ServiceName, services.Sources).Register(api)

	addr := ":" + cfg.Port
	log.Printf("[%s] starting HTTP server on %s", cfg.ServiceName, addr)
	return router.Run(addr)
}
