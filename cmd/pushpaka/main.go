package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	workerapp "github.com/yourusername/pushpaka-worker/app"
	apiapp "github.com/yourusername/pushpaka/app"
)

// setIfEmpty sets an environment variable only when it hasn't been set yet.
func setIfEmpty(key, val string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, val)
	}
}

func main() {
	dev := flag.Bool("dev", false,
		"Development mode: use SQLite (no Postgres/Redis required), run API only, enable debug logging")
	flag.Parse()

	if *dev {
		setIfEmpty("APP_ENV", "development")
		setIfEmpty("DATABASE_DRIVER", "sqlite")
		setIfEmpty("DATABASE_URL", "pushpaka-dev.db")
		setIfEmpty("PUSHPAKA_COMPONENT", "api")
		setIfEmpty("LOG_LEVEL", "debug")
		setIfEmpty("PORT", "8080")
		setIfEmpty("JWT_SECRET", "dev-secret-change-in-production")
		// Leave REDIS_URL empty so Redis is skipped (deploy triggers show a clear error)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("APP_ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	component := os.Getenv("PUSHPAKA_COMPONENT")
	if component == "" {
		component = "all"
	}

	log.Info().
		Str("component", component).
		Str("version", "v1.0.0").
		Msg("Pushpaka starting")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var runErr error

	switch component {
	case "api":
		runErr = apiapp.Run(ctx)

	case "worker":
		runErr = workerapp.Run(ctx)

	case "all":
		errCh := make(chan error, 2)
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := apiapp.Run(ctx); e != nil {
				errCh <- fmt.Errorf("api: %w", e)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := workerapp.Run(ctx); e != nil {
				errCh <- fmt.Errorf("worker: %w", e)
			}
		}()

		// If either component fails, cancel the other.
		go func() {
			for e := range errCh {
				log.Error().Err(e).Msg("component error, initiating shutdown")
				cancel()
			}
		}()

		wg.Wait()
		close(errCh)

	default:
		log.Fatal().
			Str("component", component).
			Str("valid", "api | worker | all").
			Msg("unknown PUSHPAKA_COMPONENT")
	}

	if runErr != nil {
		log.Fatal().Err(runErr).Msg("pushpaka exited with error")
	}
}
