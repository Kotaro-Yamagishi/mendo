package kitchen

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"

	"mendo/internal/domain"
	"mendo/internal/domain/contract"
	kitchendomain "mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
)

// StartCookingUsecase は調理開始のユースケース。
// OrderConfirmed イベントを購読して、厨房に調理タスクを作る。
type StartCookingUsecase struct {
	kitchenReader kitchendomain.Reader
	kitchenWriter kitchendomain.Writer
	publisher     domain.EventPublisher
	kitchenID     kitchendomain.KitchenID // この店舗の厨房ID
}

func NewStartCookingUsecase(kr kitchendomain.Reader, kw kitchendomain.Writer, pub domain.EventPublisher, id kitchendomain.KitchenID) *StartCookingUsecase {
	return &StartCookingUsecase{kitchenReader: kr, kitchenWriter: kw, publisher: pub, kitchenID: id}
}

// HandleOrderConfirmedPublic は公開イベントを受け取って調理タスクを作る。
// Order BC の内部イベント構造を知らない。公開イベントに載った調理情報だけ使う。
func (uc *StartCookingUsecase) HandleOrderConfirmedPublic(ctx context.Context, event contract.OrderConfirmedPublic) error {
	ctx, span := otel.Tracer("mendo").Start(ctx, "StartCooking.HandleOrderConfirmedPublic")
	defer span.End()

	// 1. 厨房集約をロード
	k, err := uc.kitchenReader.FindByID(ctx, uc.kitchenID)
	if err != nil {
		return err
	}

	// 2. 公開イベントから CookingInstruction を組み立てる
	instructions := make([]kitchendomain.CookingInstruction, len(event.Items))
	for i, item := range event.Items {
		instructions[i] = kitchendomain.CookingInstruction{
			MenuName: item.MenuName,
			Toppings: item.Toppings,
			Hardness: item.Hardness,
		}
	}

	// 3. 調理タスクを追加（業務ルールは集約の中）
	if err := k.AddCookingTask(order.OrderID(event.OrderID), instructions); err != nil {
		// フル稼働 → CookingRejected を Publish して補償アクションに委ねる
		if pubErr := uc.publisher.Publish(ctx, k.DomainEvents()...); pubErr != nil {
			return pubErr
		}
		slog.InfoContext(ctx, "saga: cooking task rejected, compensation event published",
			slog.String("order_id", event.OrderID),
		)
		return nil // subscriber 全体のエラーにはしない
	}

	// 4. 保存
	if err := uc.kitchenWriter.Save(ctx, k); err != nil {
		return err
	}

	slog.InfoContext(ctx, "cooking started", slog.String("order_id", event.OrderID), slog.String("kitchen_id", string(uc.kitchenID)))
	return nil
}
