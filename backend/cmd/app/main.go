package main

import (
	"backend/internal/config"
	"backend/internal/embedding_client"
	"backend/internal/handler"
	"backend/internal/repository"
	"backend/internal/server"
	"backend/internal/service"
	"backend/pkg/logger"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var cfgPath string
	var dbEnvPath string
	var backendEnvPath string

	flag.StringVar(&cfgPath, "cfg", "configs/local", "backend config path")
	flag.StringVar(&dbEnvPath, "db-env", "../.env.db", "path to db .env file")
	flag.StringVar(&backendEnvPath, "backend-env", "../.env.backend", "path to backend .env file")

	flag.Parse()

	cfg, err := config.Init(cfgPath, dbEnvPath, backendEnvPath)
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}

	log := logger.New(cfg.LogLevel)

	ctx, cancel := context.WithCancel(context.Background())

	pool, err := pgxpool.New(ctx, cfg.Db.GetUrl())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to gpxpool")
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to ping pg")
	}

	tokenAuth := jwtauth.New("HS256", []byte(cfg.JWT.Secret), nil)

	repo := repository.NewPostgres(pool, &log)

	embeddingClient := embedding_client.NewClient(cfg.Embedding.GetUrl())

	service := service.New(
		repo,
		tokenAuth,
		embeddingClient,
		&log,
	)

	handler := handler.NewHandler(
		cfg.Handler,
		service,
		tokenAuth,
		&log,
	).Init()

	server := server.NewServer(cfg.Server, handler)

	errChan := make(chan error, 2)
	go func() {
		log.Info().Int("port", cfg.Server.Port).Msg("server listening")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-quit:
		log.Info().Msg("Shutdown signal received...")
	case err := <-errChan:
		log.Info().Err(err).Msg("Error received")
	}

	cancel()
	pool.Close()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Info().Err(err).Msg("HTTP server Shutdown")
	}

	log.Info().Msg("Shutdown complete")
}
// Test rebuild
