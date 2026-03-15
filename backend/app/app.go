// Package app exposes the Pushpaka API server as a callable function,
// allowing it to be embedded in the combined pushpaka binary.
package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/yourusername/pushpaka/internal/config"
	"github.com/yourusername/pushpaka/internal/database"
	"github.com/yourusername/pushpaka/internal/router"
)

// Run starts the Pushpaka API server and blocks until ctx is cancelled.
func Run(ctx context.Context) error {
	cfg := config.Load()

	db, err := database.New(cfg.DatabaseDriver, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer db.Close()

	// Redis is optional: skipped when REDIS_URL is empty (e.g. dev/sqlite mode).
	var rdb *redis.Client
	if cfg.RedisURL != "" {
		rdb, err = database.NewRedis(cfg.RedisURL)
		if err != nil {
			log.Warn().Err(err).Msg("redis unavailable – deployment triggers disabled")
		} else {
			defer rdb.Close()
		}
	} else {
		log.Warn().Msg("REDIS_URL not set – deployment triggers disabled")
	}

	r := router.New(cfg, db, rdb)

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
