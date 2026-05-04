package main

import (
	"context"
	"log"

	"mendo/internal/di"
	"mendo/internal/domain/kitchen"
	"mendo/internal/infrastructure/eventbus"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/server"
)

func main() {
	kitchenID := kitchen.KitchenID("kitchen-001")
	dlqStore := repository.NewInMemoryDLQ()
	bus := eventbus.NewWatermillEventBus(dlqStore, 3)

	app, err := di.InitializeApp(kitchenID, bus, dlqStore)
	if err != nil {
		log.Fatal(err)
	}

	app.OutboxRelay.Start(context.Background())

	server.Run(app, bus)
}
