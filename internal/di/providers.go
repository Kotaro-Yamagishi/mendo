package di

import (
	"time"

	"github.com/google/wire"

	dlqcommand "mendo/internal/application/command/dlq"
	kitchencommand "mendo/internal/application/command/kitchen"
	menucommand "mendo/internal/application/command/menu"
	ordercommand "mendo/internal/application/command/order"
	socommand "mendo/internal/application/command/specialorder"
	dlqquery "mendo/internal/application/query/dlq"
	orderquery "mendo/internal/application/query/order"
	"mendo/internal/domain"
	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/service"
	infraoutbox "mendo/internal/infrastructure/outbox"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/interface/handler"
)

// --- インフラ層 ---

//nolint:unused // Wire で使用
var providerSet = wire.NewSet(
	provideMenuRepository,
	provideKitchenRepository,
	provideEventStore,
	provideOrderBoard,
	provideOrderStateStore,
	provideOutbox,
	provideRelayService,
	provideSpecialOrderRepository,
	provideTransactionManager,
	wire.Bind(new(domain.DeadLetterQueue), new(*repository.InMemoryDLQ)),
	wire.Bind(new(domain.TransactionManager), new(*repository.InMemoryTransactionManager)),
)

func provideMenuRepository() *repository.InMemoryMenuRepository {
	return repository.NewInMemoryMenuRepository()
}

func provideKitchenRepository() *repository.InMemoryKitchenRepository {
	return repository.NewInMemoryKitchenRepository()
}

func provideEventStore() *repository.InMemoryEventStore {
	return repository.NewInMemoryEventStore()
}

func provideOrderBoard() *repository.OrderBoardProjection {
	return repository.NewOrderBoardProjection()
}

func provideOrderStateStore() *repository.InMemoryOrderStateStore {
	return repository.NewInMemoryOrderStateStore()
}

func provideOutbox() *repository.InMemoryOutbox {
	return repository.NewInMemoryOutbox()
}

func provideTransactionManager() *repository.InMemoryTransactionManager {
	return repository.NewInMemoryTransactionManager()
}

func provideRelayService(ob *repository.InMemoryOutbox, pub domain.EventPublisher) *infraoutbox.RelayService {
	return infraoutbox.NewRelayService(ob, pub, 500*time.Millisecond)
}

// --- ドメインサービス ---

//nolint:unused // Wire で使用
var domainServiceSet = wire.NewSet(
	provideDomainServices,
)

func provideDomainServices(
	orderStore *repository.InMemoryOrderStateStore,
	kitchenRepo *repository.InMemoryKitchenRepository,
) *service.WaitTimeCalculator {
	return service.NewWaitTimeCalculator(orderStore, kitchenRepo)
}

// --- アプリケーション層 ---

//nolint:unused // Wire で使用
var usecaseSet = wire.NewSet(
	provideCreateOrderUsecase,
	provideConfirmOrderUsecase,
	provideCancelOrderUsecase,
	provideCompleteCookingUsecase,
	provideEstimateWaitTimeUsecase,
	provideSoldOutMenuUsecase,
	provideStartCookingUsecase,
	provideListOrdersUsecase,
	provideRetryDLQUsecase,
	provideListDLQHandler,
	provideCreateSpecialOrderUsecase,
	provideApproveSpecialOrderUsecase,
	provideRejectSpecialOrderUsecase,
	provideResubmitSpecialOrderUsecase,
)

func provideCreateOrderUsecase(
	es domain.EventStore,
	ob *repository.InMemoryOutbox,
) *ordercommand.CreateOrderUsecase {
	return ordercommand.NewCreateOrderUsecase(es, ob)
}

func provideConfirmOrderUsecase(
	es domain.EventStore,
	ob *repository.InMemoryOutbox,
) *ordercommand.ConfirmOrderUsecase {
	return ordercommand.NewConfirmOrderUsecase(es, ob)
}

func provideCancelOrderUsecase(
	es domain.EventStore,
	ob *repository.InMemoryOutbox,
) *ordercommand.CancelOrderUsecase {
	return ordercommand.NewCancelOrderUsecase(es, ob)
}

func provideCompleteCookingUsecase(
	kitchenRepo *repository.InMemoryKitchenRepository,
	ob *repository.InMemoryOutbox,
	tx domain.TransactionManager,
	kitchenID kitchen.KitchenID,
) *kitchencommand.CompleteCookingUsecase {
	return kitchencommand.NewCompleteCookingUsecase(kitchenRepo, kitchenRepo, ob, tx, kitchenID)
}

func provideEstimateWaitTimeUsecase(
	calc *service.WaitTimeCalculator,
	kitchenID kitchen.KitchenID,
) *orderquery.EstimateWaitTimeUsecase {
	return orderquery.NewEstimateWaitTimeUsecase(calc, kitchenID)
}

func provideSoldOutMenuUsecase(
	menuRepo *repository.InMemoryMenuRepository,
) *menucommand.SoldOutMenuUsecase {
	return menucommand.NewSoldOutMenuUsecase(menuRepo, menuRepo)
}

func provideStartCookingUsecase(
	kitchenRepo *repository.InMemoryKitchenRepository,
	pub domain.EventPublisher,
	kitchenID kitchen.KitchenID,
) *kitchencommand.StartCookingUsecase {
	return kitchencommand.NewStartCookingUsecase(kitchenRepo, kitchenRepo, pub, kitchenID)
}

func provideListOrdersUsecase(store *repository.InMemoryOrderStateStore) *orderquery.ListOrdersUsecase {
	return orderquery.NewListOrdersUsecase(store)
}

func provideRetryDLQUsecase(d domain.DeadLetterQueue, pub domain.EventPublisher) *dlqcommand.RetryDLQUsecase {
	return dlqcommand.NewRetryDLQUsecase(d, pub)
}

func provideListDLQHandler(d domain.DeadLetterQueue) *dlqquery.ListDLQHandler {
	return dlqquery.NewListDLQHandler(d)
}

// --- インターフェース層 ---

//nolint:unused // Wire で使用
var handlerSet = wire.NewSet(
	handler.NewOrderHandler,
	handler.NewKitchenHandler,
	handler.NewMenuHandler,
	handler.NewDLQHandler,
	handler.NewSpecialOrderHandler,
)

// --- 特別注文（プロセスマネージャー）---

func provideSpecialOrderRepository() *repository.InMemorySpecialOrderRepository {
	return repository.NewInMemorySpecialOrderRepository()
}

func provideCreateSpecialOrderUsecase(repo *repository.InMemorySpecialOrderRepository, pub domain.EventPublisher) *socommand.CreateSpecialOrderUsecase {
	return socommand.NewCreateSpecialOrderUsecase(repo, pub)
}

func provideApproveSpecialOrderUsecase(repo *repository.InMemorySpecialOrderRepository, pub domain.EventPublisher) *socommand.ApproveSpecialOrderUsecase {
	return socommand.NewApproveSpecialOrderUsecase(repo, repo, pub)
}

func provideRejectSpecialOrderUsecase(repo *repository.InMemorySpecialOrderRepository, pub domain.EventPublisher) *socommand.RejectSpecialOrderUsecase {
	return socommand.NewRejectSpecialOrderUsecase(repo, repo, pub)
}

func provideResubmitSpecialOrderUsecase(repo *repository.InMemorySpecialOrderRepository, pub domain.EventPublisher) *socommand.ResubmitSpecialOrderUsecase {
	return socommand.NewResubmitSpecialOrderUsecase(repo, repo, pub)
}

