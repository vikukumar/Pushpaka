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
	"github.com/vikukumar/Pushpaka/queue"
)

func main() {
	dev := flag.Bool("dev", false, "dev mode: use SQLite + embedded worker (no Postgres/Redis required)")
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	component := os.Getenv("PUSHPAKA_COMPONENT")
	if component == "" {
		component = "all"
	}

	var wg sync.WaitGroup

	switch component {
	case "all":
		// API + embedded worker in one process, connected by a fast in-process channel.
		// Works in both -dev mode (SQLite) and production (Postgres).
		// Redis is NOT used for job routing in this mode — the in-process channel
		// handles all dispatch and gives the API live worker/job counts.
		// For horizontal worker scaling, use PUSHPAKA_COMPONENT=api + PUSHPAKA_COMPONENT=worker.
		//
		// IMPORTANT: open ONE shared database pool and hand it to BOTH components.
		// This prevents SQLITE_BUSY_SNAPSHOT (error 261) in WAL mode: that error
		// occurs when two separate connection pools hold overlapping read snapshots
		// and one tries to write.  A single pool serialises all access.
		sharedDB, err := backendApp.OpenDB(os.Getenv("DATABASE_DRIVER"), os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open shared database")
		}
		defer sharedDB.Close()

		q := queue.New(100)
		log.Info().Bool("dev", *dev).Msg("all-in-one mode: in-process queue active, embedded workers starting")

		wg.Add(1)
		go func() {
			defer wg.Done()
			opts := backendApp.RunOptions{InProcessQueue: q, DB: sharedDB}
			if err := backendApp.RunWithOptions(ctx, opts); err != nil {
				log.Error().Err(err).Msg("API server error")
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := workerApp.RunInProcessWithDB(ctx, q.Chan(), q, sharedDB); err != nil {
				log.Error().Err(err).Msg("embedded worker error")
			}
		}()

	case "api":
		// API only — expects external worker processes reading from Redis.
		log.Info().Msg("api-only mode: external workers must connect via Redis")
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := backendApp.Run(ctx); err != nil {
				log.Error().Err(err).Msg("API server error")
			}
		}()

	case "worker":
		// Worker only — reads jobs from Redis, pairs with PUSHPAKA_COMPONENT=api.
		log.Info().Msg("worker-only mode: connecting to Redis queue")
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := workerApp.Run(ctx); err != nil {
				log.Error().Err(err).Msg("worker error")
			}
		}()

	default:
		log.Fatal().Str("component", component).Msg("unknown PUSHPAKA_COMPONENT; valid values: all, api, worker")
	}

	wg.Wait()
	log.Info().Msg("pushpaka stopped")
}

func setIfEmpty(key, val string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, val)
	}
}
