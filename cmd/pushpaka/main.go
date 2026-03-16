// Command pushpaka is the combined Pushpaka binary.
// Use -dev flag for local SQLite development (no Postgres/Redis required).
// Use PUSHPAKA_COMPONENT={api|worker|all} to select which component to run.
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	workerApp "github.com/vikukumar/Pushpaka-worker/app"
	backendApp "github.com/vikukumar/Pushpaka/app"
)

func main() {
	dev := flag.Bool("dev", false, "dev mode: use SQLite + no Redis required")
	flag.Parse()

	if *dev {
		setIfEmpty("DATABASE_DRIVER", "sqlite")
		setIfEmpty("DATABASE_URL", "pushpaka-dev.db")
		setIfEmpty("REDIS_URL", "")
		setIfEmpty("JWT_SECRET", "dev-secret-change-in-production")
		setIfEmpty("APP_ENV", "development")
		setIfEmpty("PORT", "8080")
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("APP_ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	component := os.Getenv("PUSHPAKA_COMPONENT")
	if component == "" {
		if *dev {
			component = "api" // worker needs Redis; skip in dev mode
		} else {
			component = "all"
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup

	if component == "api" || component == "all" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := backendApp.Run(ctx); err != nil {
				log.Error().Err(err).Msg("API server error")
			}
		}()
	}

	if component == "worker" || component == "all" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := workerApp.Run(ctx); err != nil {
				log.Error().Err(err).Msg("worker error")
			}
		}()
	}

	wg.Wait()
	log.Info().Msg("pushpaka stopped")
}

func setIfEmpty(key, val string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, val)
	}
}
