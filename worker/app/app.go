// Package app exposes the Pushpaka build-worker pool as a callable function,
// allowing it to be embedded in the combined pushpaka binary.
package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/yourusername/pushpaka-worker/internal/config"
	"github.com/yourusername/pushpaka-worker/internal/database"
	"github.com/yourusername/pushpaka-worker/internal/worker"
)

// Run starts the Pushpaka build worker pool and blocks until ctx is cancelled.
func Run(ctx context.Context) error {
	cfg := config.Load()

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
