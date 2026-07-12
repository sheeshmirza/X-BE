package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"X-BE/internal/config"
	"X-BE/internal/database"
	"X-BE/internal/handlers"
	"X-BE/internal/repository"
	"X-BE/internal/router"
	"X-BE/internal/service"
)

func main() {
	cfg := config.Load()
	mongoClient, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect mongo: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if disconnectErr := mongoClient.Disconnect(ctx); disconnectErr != nil {
			log.Printf("Mongo disconnect error: %v", disconnectErr)
		}
	}()
	repos := repository.NewRepositories(mongoClient.Database(cfg.MongoDBName), cfg)
	services := service.NewServices(repos)
	handlers := handlers.NewHandlers(services)
	httpRouter := router.New(cfg, handlers)
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	go func() {
		log.Printf("X-BE listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
	}
}
