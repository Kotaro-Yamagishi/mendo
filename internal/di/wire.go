//go:build wireinject
// +build wireinject

package di

import (
	"log/slog"

	"github.com/google/wire"

	kitchencommand "mendo/internal/application/command/kitchen"
	ordercommand "mendo/internal/application/command/order"
	"mendo/internal/domain"
	"mendo/internal/domain/kitchen"
	"mendo/internal/infrastructure/importworker"
	infraoutbox "mendo/internal/infrastructure/outbox"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/interface/handler"
)

// App は全ハンドラとイベント購読者を保持する構造体。
type App struct {
	OrderHandler        *handler.OrderHandler
	KitchenHandler      *handler.KitchenHandler
	MenuHandler         *handler.MenuHandler
	DLQHandler          *handler.DLQHandler
	SpecialOrderHandler *handler.SpecialOrderHandler
	ImportHandler       *handler.ImportHandler
	StartCookingUC      *kitchencommand.StartCookingUsecase
	CancelOrderUC       *ordercommand.CancelOrderUsecase
	OrderBoard          *repository.OrderBoardProjection
	OrderStateStore     *repository.InMemoryOrderStateStore
	OutboxRelay         *infraoutbox.RelayService
	ImportWorker        *importworker.Worker
	Logger              *slog.Logger
}

func InitializeApp(kitchenID kitchen.KitchenID, eventBus domain.EventPublisher, dlqStore *repository.InMemoryDLQ, logger *slog.Logger) (*App, error) {
	wire.Build(
		// インフラ層: リポジトリ実装
		providerSet,

		// ドメインサービス
		domainServiceSet,

		// アプリケーション層: ユースケース
		usecaseSet,

		// インターフェース層: ハンドラ
		handlerSet,

		// App を組み立て
		wire.Struct(new(App), "*"),
	)
	return nil, nil
}
