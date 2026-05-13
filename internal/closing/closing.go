package closing

import (
	"context"
	"log/slog"
)

// CloseShop は閉店処理を実行するトランザクションスクリプト。
//
// ドメインモデルやアクティブレコードと異なり:
//   - 構造体もモデルもない。関数1つで完結
//   - 業務ルールなし。「未完了の注文を全部キャンセルする」だけ
//   - 本番では SQL 1本で済む:
//     db.Exec("UPDATE orders SET status = 'canceled' WHERE status IN ('pending','confirmed')")
//
// 第10章: 補完領域 → トランザクションスクリプト → レイヤードアーキテクチャ
func CloseShop(ctx context.Context, orderIDs []string) (int, error) {
	canceledCount := 0
	for _, orderID := range orderIDs {
		slog.InfoContext(ctx, "close shop: order canceled", slog.String("order_id", orderID))
		canceledCount++
	}
	slog.InfoContext(ctx, "close shop completed", slog.Int("canceled_count", canceledCount))
	return canceledCount, nil
}
