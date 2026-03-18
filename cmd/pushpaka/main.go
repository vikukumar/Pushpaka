// Command pushpaka is the combined Pushpaka binary.
// Use -dev flag for local SQLite development (no Postgres/Redis required).
// Use PUSHPAKA_COMPONENT={api|worker|all} to select which component to run.
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

	workerApp "github.com/vikukumar/Pushpaka-worker/app"
	backendApp "github.com/vikukumar/Pushpaka/app"
	"github.com/vikukumar/Pushpaka/queue"
)

var version = "dev"

func main() {
	dev := flag.Bool("dev", false, "dev mode: use SQLite + embedded worker (no Postgres/Redis required)")
	configPath := flag.String("config", os.Getenv("PUSHPAKA_CONFIG_FILE"), "path to config.yaml")
	mode := flag.String("mode", "", "config mode: development, staging, production")
	flag.Parse()

	if *configPath != "" {
		selectedMode := *mode
		if selectedMode == "" {
			if *dev {
				selectedMode = "development"
			} else if envMode := os.Getenv("APP_ENV"); envMode != "" {
				selectedMode = envMode
			} else {
				selectedMode = "production"
			}
		}
		if err := applyConfigFile(*configPath, selectedMode); err != nil {
			log.Fatal().Err(err).Msg("failed to load config file")
		}
		log.Info().Str("config_file", *configPath).Str("mode", normalizeConfigMode(selectedMode)).Msg("loaded config file")
	}

	if *dev {
		os.Setenv("DATABASE_DRIVER", "sqlite")
		os.Setenv("DATABASE_URL", "pushpaka-dev.db")
		os.Setenv("REDIS_URL", "")
		os.Setenv("REDIS_ENABLED", "false")
		os.Setenv("APP_ENV", "development")
		os.Setenv("PORT", "8080")
		setIfEmpty("JWT_SECRET", "dev-secret-change-in-production")
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

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output(), "")
		fmt.Fprintln(flag.CommandLine.Output(), "The config file, when provided, overrides environment-derived values for the selected mode.")
	}
}
