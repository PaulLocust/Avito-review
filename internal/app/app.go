// app.go
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PaulLocust/Avito-review/config"
	"github.com/PaulLocust/Avito-review/internal/controller/http"
	"github.com/PaulLocust/Avito-review/internal/repository/postgresql"
	"github.com/PaulLocust/Avito-review/internal/usecase"
	"github.com/PaulLocust/Avito-review/pkg/httpserver"
	"github.com/PaulLocust/Avito-review/pkg/logger"
	"github.com/PaulLocust/Avito-review/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)
	l.Info("Starting application...")

	// Repository
	l.Info("Connecting to database...")
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()
	l.Info("Database connected successfully")

	// Use-Cases с реальными зависимостями
	l.Info("Initializing use cases...")

	// Инициализируем репозитории с логгером
	teamRepo := postgresql.NewTeamRepository(pg.Pool, l)
	userRepo := postgresql.NewUserRepository(pg.Pool, l)
	prRepo := postgresql.NewPRRepository(pg.Pool, l)

	useCases := usecase.NewUseCases(teamRepo, userRepo, prRepo, l)
	l.Info("Use cases initialized successfully")

	// HTTP Router (net/http)
	l.Info("Setting up HTTP router...")
	handler := http.NewRouter(cfg, l, useCases)

	// HTTP Server
	l.Info("Creating HTTP server...")
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Start server
	l.Info("Starting HTTP server on port %s...", cfg.HTTP.Port)
	httpServer.Start()

	// Даем серверу время на запуск
	l.Info("Waiting for server to start...")
	time.Sleep(500 * time.Millisecond)

	// Проверяем что сервер запустился
	select {
	case err := <-httpServer.Notify():
		l.Fatal(fmt.Errorf("server failed to start: %w", err))
	default:
		l.Info("HTTP server started successfully on port %s", cfg.HTTP.Port)
		l.Info("Application is running. Press Ctrl+C to stop.")
	}

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("Shutdown signal received: %s", s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("server error: %w", err))
	}

	// Shutdown
	l.Info("Shutting down server...")
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("shutdown error: %w", err))
	}
	l.Info("Application stopped")
}