// Package app exposes the Pushpaka build-worker pool as a callable function,
// allowing it to be embedded in the combined pushpaka binary.
package app

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/database"
	"github.com/vikukumar/Pushpaka/pkg/models"

	"github.com/vikukumar/Pushpaka/worker/internal/client"
	"github.com/vikukumar/Pushpaka/worker/internal/config"
	"github.com/vikukumar/Pushpaka/worker/internal/worker"
)

// StatsReporter is implemented by the in-process queue to track worker and job lifecycle events.
// All methods must be safe for concurrent use. Pass nil to disable reporting.
type StatsReporter interface {
	WorkerStarted()
	WorkerStopped()
	JobStarted()
	JobFinished()
}

// RunOptions defines the configuration for a remote worker node
type RunOptions struct {
	Mode      string
	ServerURL string
	ZonePAT   string
}

// Run starts the Pushpaka worker pool and blocks until ctx is cancelled.
func Run(ctx context.Context, opts RunOptions) error {
	cfg := config.Load()

	if err := cfg.EnsureDirs(); err != nil {
		return fmt.Errorf("dirs: %w", err)
	}
	log.Info().
		Str("clone_dir", cfg.CloneDir).
		Str("deploy_dir", cfg.DeployDir).
		Msg("build directories ready")

	log.Info().
		Str("mode", opts.Mode).
		Str("server", opts.ServerURL).
		Msg("Initializing Pushpaka worker node")

	var db *gorm.DB
	var err error
	if opts.Mode == "hybrid" {
		db, err = basemodel.Connect(cfg.DatabaseDriver, cfg.DatabaseURL, "production")
	} else if opts.Mode == "vaahan" {
		log.Info().Msg("Vaahan mode: using embedded pure-Go SQLite for lightning-fast serverless tracking")
		db, err = basemodel.Connect("sqlite", filepath.Join(cfg.DeployDir, "vaahan.db"), "production")
	} else {
		return fmt.Errorf("unknown worker mode: %s (must be hybrid or vaahan)", opts.Mode)
	}

	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	// In Vaahan mode, ensure the schema is initialized if brand new.
	if opts.Mode == "vaahan" {
		_ = db.AutoMigrate(&models.Deployment{}, &models.DeploymentLog{})
	}

	// In remote modes (vaahan/hybrid), the worker bypasses Redis and receives deployment tasks
	// directly over the Yamux multiplexed websocket from the main API!
	if opts.Mode == "vaahan" || opts.Mode == "hybrid" {
		deployCh := make(chan []byte, 100)

		nodeClient := client.NewWorkerClient(opts.ServerURL, opts.ZonePAT, db, cfg, deployCh)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			nodeClient.Start(ctx)
		}()

		log.Info().Msg("Starting in-process deployment workers bound to Yamux tunnel")
		if err := runInProcess(ctx, deployCh, nil, db, cfg); err != nil {
			return err
		}

		wg.Wait()
		return nil
	}

	// Legacy or standalone mode: relies on Redis queue
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

	log.Info().Int("build_workers", cfg.BuildWorkers).Msg("Pushpaka redis workers started")
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

	db, err := basemodel.Connect(cfg.DatabaseDriver, cfg.DatabaseURL, "production")
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	return runInProcess(ctx, ch, reporter, db, cfg)
}

// RunInProcessWithDB is identical to RunInProcess but uses a pre-opened
// database connection instead of opening a new one.  The caller retains
// ownership of db and must close it after this function returns.
// Use this in all-in-one mode to share ONE SQLite pool between the API
// and the embedded worker, preventing SQLITE_BUSY_SNAPSHOT (error 261).
func RunInProcessWithDB(ctx context.Context, ch <-chan []byte, reporter StatsReporter, db *gorm.DB) error {
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

func runInProcess(ctx context.Context, ch <-chan []byte, reporter StatsReporter, db *gorm.DB, cfg *config.Config) error {
	var wg sync.WaitGroup

	if reporter != nil {
		for i := 0; i < cfg.BuildWorkers; i++ {
			reporter.WorkerStarted()
		}
	}

	for i := 0; i < cfg.BuildWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if reporter != nil {
				defer reporter.WorkerStopped()
			}
			w := worker.NewBuildWorker(id, db, nil, cfg)
			w.RunInProcess(ctx, ch, reporter)
		}(i)
	}

	log.Info().Int("build_workers", cfg.BuildWorkers).Msg("in-process workers started")
	wg.Wait()
	log.Info().Msg("in-process workers stopped")
	return nil
}
