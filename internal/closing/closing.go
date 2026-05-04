package closing

import (
	"fmt"
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
func CloseShop(orderIDs []string) (int, error) {
	canceledCount := 0
	for _, orderID := range orderIDs {
		fmt.Printf("[CloseShop] 注文 %s をキャンセル\n", orderID)
		canceledCount++
	}
	fmt.Printf("[CloseShop] 閉店処理完了。%d 件の注文をキャンセル\n", canceledCount)
	return canceledCount, nil
}
