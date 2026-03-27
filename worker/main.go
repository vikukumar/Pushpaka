package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/vikukumar/pushpaka/worker/app"
)

var version = "dev"

func main() {
	mode := flag.String("mode", "hybrid", "worker run mode: vaahan (serverless) or hybrid")
	serverURL := flag.String("server", "ws://127.0.0.1:8081", "Pushpaka API Management Server URL")
	zonePAT := flag.String("zone-pat", "", "Personal Access Token for the Zone")
	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("APP_ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	opts := app.RunOptions{
		Mode:      *mode,
		ServerURL: *serverURL,
		ZonePAT:   *zonePAT,
	}

	if err := app.Run(ctx, opts); err != nil {
		log.Fatal().Err(err).Msg("worker exited with error")
	}
}
