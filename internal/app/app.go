package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-co-op/gocron"

	"github.com/realPointer/segments/config"
	v1 "github.com/realPointer/segments/internal/controller/http/v1"
	"github.com/realPointer/segments/internal/repo"
	"github.com/realPointer/segments/internal/service"
	"github.com/realPointer/segments/internal/ydisk/ydisk"
	"github.com/realPointer/segments/pkg/httpserver"
	"github.com/realPointer/segments/pkg/logger"
	"github.com/realPointer/segments/pkg/postgres"
)

func Run() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Logger
	l := logger.New(cfg.Log.Level)
	l.Info("Config and logger initialized")

	// Postgres
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	err = pg.Pool.Ping(context.Background())
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - pg.Pool.Ping: %w", err))
	}

	// Repositories
	repositories := repo.NewRepositories(pg)

	// Services dependencies
	deps := service.ServicesDependencies{
		Repos:      repositories,
		YandexDisk: ydisk.NewYandexDisk(cfg.WebAPI.YandexToken),
	}
	services := service.NewServices(deps)

	// GoCron
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Minute().Do(services.Scheduler.DeleteExpiredRows, context.Background())
	s.StartAsync()

	// HTTP Server
	handler := chi.NewRouter()
	v1.NewRouter(handler, l, services)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
