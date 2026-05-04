package domain

import "context"

// TransactionManager はトランザクション境界を管理する。
// 複数の集約やリポジトリの操作を1つのトランザクションで束ねる。
//
// 用途:
//   - 集約の保存と Outbox へのイベント保存を同一トランザクションで行う
//   - 同じBC内の複数集約を同一トランザクションで保存する
//
// 本番では DB のトランザクション（BEGIN/COMMIT/ROLLBACK）で実装する。
// InMemory 実装ではロック制御で擬似的に再現する。
type TransactionManager interface {
	// Do はコールバック関数をトランザクション内で実行する。
	// コールバックがエラーを返した場合、トランザクションはロールバックされる。
	// コールバックが nil を返した場合、トランザクションはコミットされる。
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
