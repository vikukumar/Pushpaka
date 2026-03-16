// Package app exposes the Pushpaka API server as a callable function,
// allowing it to be embedded in the combined pushpaka binary.
package app

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/vikukumar/Pushpaka/internal/config"
	"github.com/vikukumar/Pushpaka/internal/database"
	"github.com/vikukumar/Pushpaka/internal/router"
	"github.com/vikukumar/Pushpaka/queue"
	"github.com/vikukumar/Pushpaka/ui"
)

// RunOptions configures optional behaviour for Run.
type RunOptions struct {
	// InProcessQueue, when non-nil, is used instead of Redis for deployment jobs.
	// Intended for dev mode where the embedded worker reads from the same queue.
	InProcessQueue *queue.InProcess
}

// Run starts the Pushpaka API server with default options and blocks until ctx is cancelled.
func Run(ctx context.Context) error {
	return RunWithOptions(ctx, RunOptions{})
}

// RunWithOptions starts the Pushpaka API server with the supplied options.
func RunWithOptions(ctx context.Context, opts RunOptions) error {
	cfg := config.Load()

	db, err := database.New(cfg.DatabaseDriver, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer db.Close()

	// Redis is optional: skipped when REDIS_URL is empty or an in-process queue is used.
	var rdb *redis.Client
	if opts.InProcessQueue != nil {
		log.Info().Msg("using in-process job queue (dev mode)")
	} else if cfg.RedisURL != "" {
		rdb, err = database.NewRedis(cfg.RedisURL)
		if err != nil {
			log.Warn().Err(err).Msg("redis unavailable - deployment triggers disabled")
		} else {
			defer rdb.Close()
		}
	} else {
		log.Warn().Msg("REDIS_URL not set and no in-process queue — deployment triggers disabled")
	}

	// Detect whether the frontend was compiled into the binary.
	// In dev mode ui/dist only contains a placeholder, so uiFS stays nil.
	var uiFS fs.FS
	if _, ferr := ui.FS.Open("dist/index.html"); ferr == nil {
		if sub, serr := fs.Sub(ui.FS, "dist"); serr == nil {
			uiFS = sub
			log.Info().Msg("serving embedded frontend")
		}
	}

	r := router.New(cfg, db, rdb, uiFS, opts.InProcessQueue)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info().Str("port", cfg.Port).Str("version", "v1.0.0").Msg("Pushpaka API starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Info().Msg("shutting down API server...")
	return srv.Shutdown(shutCtx)
}
