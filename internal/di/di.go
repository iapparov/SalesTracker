package di

import (
	"context"
	"fmt"
	wbgin "github.com/wb-go/wbf/ginext"
	"go.uber.org/fx"
	"log"
	"net/http"
	"salestracker/internal/config"
	"salestracker/internal/storage/postgres"
	"salestracker/internal/web"
	"salestracker/internal/web/handlers"
)

func StartHTTPServer(lc fx.Lifecycle, transactionHandler *handlers.TransactionHandler, analyticsHandler *handlers.AnalyticsHandler, config *config.AppConfig) {
	router := wbgin.New(config.GinConfig.Mode)

	router.Use(wbgin.Logger(), wbgin.Recovery())
	router.Use(func(c *wbgin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	web.RegisterRoutes(router, transactionHandler, analyticsHandler)

	addres := fmt.Sprintf("%s:%d", config.ServerConfig.Host, config.ServerConfig.Port)
	server := &http.Server{
		Addr:    addres,
		Handler: router.Engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("Server started")
			go func() {
				if err := server.ListenAndServe(); err != nil {
					log.Printf("ListenAndServe error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Printf("Shutting down server...")
			return server.Close()
		},
	})
}

func ClosePostgresOnStop(lc fx.Lifecycle, postgres *postgres.Postgres) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Println("Closing Postgres connections...")
			if err := postgres.Close(); err != nil {
				log.Printf("Failed to close Postgres: %v", err)
				return err
			}
			log.Println("Postgres closed successfully")
			return nil
		},
	})
}
