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
)

func main() {
	logger := logging.NewLogger()
	slog.SetDefault(logger)

	kitchenID := kitchen.KitchenID("kitchen-001")
	dlqStore := repository.NewInMemoryDLQ()
	bus := eventbus.NewWatermillEventBus(dlqStore, 3, logger)

	app, err := di.InitializeApp(kitchenID, bus, dlqStore, logger)
	if err != nil {
		logger.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}

	app.OutboxRelay.Start(context.Background())

	server.Run(app, bus)
}
