package main

import (
	"context"
	"log/slog"
	"os"

	"mendo/internal/di"
	"mendo/internal/domain/kitchen"
	"mendo/internal/infrastructure/eventbus"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/logging"
	"mendo/internal/server"
	"mendo/internal/telemetry"
)

func main() {
	logger := logging.NewLogger()
	slog.SetDefault(logger)

	// OpenTelemetry 初期化
	shutdownTracer, err := telemetry.Init(context.Background(), "mendo")
	if err != nil {
		logger.Warn("failed to initialize OpenTelemetry, tracing disabled", "error", err)
	} else {
		defer shutdownTracer(context.Background())
	}

	kitchenID := kitchen.KitchenID("kitchen-001")
	dlqStore := repository.NewInMemoryDLQ()
	bus := eventbus.NewWatermillEventBus(dlqStore, 3, logger)

	app, err := di.InitializeApp(kitchenID, bus, dlqStore, logger)
	if err != nil {
		logger.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}

	server.Run(app, bus)
}
