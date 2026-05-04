package server

import (
	"mendo/internal/di"
	"mendo/internal/infrastructure/eventbus"
)

// RegisterSubscribers はドメインイベントの購読者を登録する。
// 集約ごとにファイルを分割して管理する。
func RegisterSubscribers(bus *eventbus.WatermillEventBus, app *di.App) {
	registerOrderSubscribers(bus, app)
	registerKitchenSubscribers(bus, app)
	registerSpecialOrderSubscribers(bus, app)
}
