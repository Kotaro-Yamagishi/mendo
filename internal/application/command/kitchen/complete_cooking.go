package kitchen

import (
	"context"

	"mendo/internal/domain"
	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
)

// CompleteCookingUsecase は調理完了のユースケース。
// 厨房スタッフが調理完了ボタンを押した時に実行される。
type CompleteCookingUsecase struct {
	kitchenReader kitchen.Reader
	kitchenWriter kitchen.Writer
	outbox        domain.Outbox
	tx            domain.TransactionManager
	kitchenID     kitchen.KitchenID
}

func NewCompleteCookingUsecase(
	kr kitchen.Reader, kw kitchen.Writer,
	ob domain.Outbox, tx domain.TransactionManager,
	id kitchen.KitchenID,
) *CompleteCookingUsecase {
	return &CompleteCookingUsecase{
		kitchenReader: kr, kitchenWriter: kw,
		outbox: ob, tx: tx, kitchenID: id,
	}
}

func (uc *CompleteCookingUsecase) Execute(ctx context.Context, orderID order.OrderID) error {
	// 1. 厨房集約をロード
	k, err := uc.kitchenReader.FindByID(ctx, uc.kitchenID)
	if err != nil {
		return err
	}

	// 2. 調理完了コマンドを実行（業務ルールは集約の中）
	if err := k.CompleteCookingTask(orderID); err != nil {
		return err // domain の AppError をそのまま返す
	}

	// 3. Kitchen の保存と Outbox へのイベント保存を同一トランザクションで実行。
	// どちらか一方が失敗したら両方ロールバックされる。
	// これにより「Kitchen は保存されたがイベントが消えた」を防ぐ。
	if err := uc.tx.Do(ctx, func(txCtx context.Context) error {
		if err := uc.kitchenWriter.Save(txCtx, k); err != nil {
			return err
		}
		if err := uc.outbox.Store(txCtx, k.DomainEvents()); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err // トランザクション内の AppError をそのまま返す
	}

	return nil
}
