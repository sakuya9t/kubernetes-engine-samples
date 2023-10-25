package main

import (
	"consumptionexp/config"
	"consumptionexp/tracker"
	"context"
	"os"

	"github.com/apsdehal/go-logger"
)

func main() {
	ctx := context.Background()
	log, _ := logger.New(config.LogLevel, os.Stdout)
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	tracker := tracker.NewMetricsTracker(ctx, cfg, log)
	defer tracker.Stop()
	tracker.Start(ctx)
	select {}
}
