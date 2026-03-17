// Package app exposes the Pushpaka build-worker pool as a callable function,
// allowing it to be embedded in the combined pushpaka binary.
package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/vikukumar/Pushpaka-worker/internal/config"
	"github.com/vikukumar/Pushpaka-worker/internal/database"
	"github.com/vikukumar/Pushpaka-worker/internal/worker"
)

// StatsReporter is implemented by the in-process queue to track worker and job lifecycle events.
// All methods must be safe for concurrent use. Pass nil to disable reporting.
type StatsReporter interface {
	WorkerStarted()
	WorkerStopped()
	JobStarted()
	JobFinished()
}

// Run starts the Pushpaka build worker pool and blocks until ctx is cancelled.
func Run(ctx context.Context) error {
	cfg := config.Load()

	if err := cfg.EnsureDirs(); err != nil {
		return fmt.Errorf("dirs: %w", err)
	}
	log.Info().
		Str("clone_dir", cfg.CloneDir).
		Str("deploy_dir", cfg.DeployDir).
		Msg("build directories ready")

	db, err := database.New(cfg.DatabaseDriver, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer db.Close()

	rdb, err := database.NewRedis(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	defer rdb.Close()

	var wg sync.WaitGroup
	for i := 0; i < cfg.BuildWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			w := worker.NewBuildWorker(id, db, rdb, cfg)
			w.Run(ctx)
		}(i)
	}

	log.Info().Int("build_workers", cfg.BuildWorkers).Msg("Pushpaka workers started")
	wg.Wait()
	log.Info().Msg("workers stopped")
	return nil
}

// RunInProcess starts the build worker pool reading from an in-process channel
// instead of Redis.  Used in dev/combined-binary mode (no Redis required).
// reporter is optional (may be nil); when provided its methods track worker
// and job lifecycle so the API layer can report live stats.
func RunInProcess(ctx context.Context, ch <-chan []byte, reporter StatsReporter) error {
	cfg := config.Load()

	if err := cfg.EnsureDirs(); err != nil {
		return fmt.Errorf("dirs: %w", err)
	}
	log.Info().
		Str("clone_dir", cfg.CloneDir).
		Str("deploy_dir", cfg.DeployDir).
		Msg("build directories ready")

	db, err := database.New(cfg.DatabaseDriver, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer db.Close()

	return runInProcess(ctx, ch, reporter, db, cfg)
}

// RunInProcessWithDB is identical to RunInProcess but uses a pre-opened
// database connection instead of opening a new one.  The caller retains
// ownership of db and must close it after this function returns.
// Use this in all-in-one mode to share ONE SQLite pool between the API
// and the embedded worker, preventing SQLITE_BUSY_SNAPSHOT (error 261).
func RunInProcessWithDB(ctx context.Context, ch <-chan []byte, reporter StatsReporter, db *sqlx.DB) error {
	cfg := config.Load()

	if err := cfg.EnsureDirs(); err != nil {
		return fmt.Errorf("dirs: %w", err)
	}
	log.Info().
		Str("clone_dir", cfg.CloneDir).
		Str("deploy_dir", cfg.DeployDir).
		Msg("build directories ready")

	return runInProcess(ctx, ch, reporter, db, cfg)
}

func runInProcess(ctx context.Context, ch <-chan []byte, reporter StatsReporter, db *sqlx.DB, cfg *config.Config) error {
	var wg sync.WaitGroup

	// Pre-register all workers synchronously *before* launching goroutines.
	// This ensures the worker count is immediately visible via /api/v1/system
	// even before the Go scheduler has run any of the worker goroutines.
	if reporter != nil {
		for i := 0; i < cfg.BuildWorkers; i++ {
			reporter.WorkerStarted()
		}
	}

	for i := 0; i < cfg.BuildWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Decrement on exit (was pre-incremented above).
			if reporter != nil {
				defer reporter.WorkerStopped()
			}
			// rdb is nil -- RunInProcess does not use Redis
			w := worker.NewBuildWorker(id, db, nil, cfg)
			w.RunInProcess(ctx, ch, reporter)
		}(i)
	}

	log.Info().Int("build_workers", cfg.BuildWorkers).Msg("in-process workers started")
	wg.Wait()
	log.Info().Msg("in-process workers stopped")
	return nil
}
