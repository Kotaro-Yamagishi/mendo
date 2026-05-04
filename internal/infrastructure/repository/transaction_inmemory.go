package repository

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryTransactionManager は学習用のインメモリトランザクション実装。
// Mutex でトランザクション中の排他制御を擬似的に再現する。
//
// 本番では *sql.Tx を使った実装に差し替える。
// 例:
//
//	func (tm *PostgresTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
//	    tx, err := tm.db.BeginTx(ctx, nil)
//	    if err != nil { return err }
//	    // tx を context に入れて、リポジトリが同じ tx を使えるようにする
//	    txCtx := context.WithValue(ctx, txKey{}, tx)
//	    if err := fn(txCtx); err != nil {
//	        tx.Rollback()
//	        return err
//	    }
//	    return tx.Commit()
//	}
type InMemoryTransactionManager struct {
	mu sync.Mutex
}

func NewInMemoryTransactionManager() *InMemoryTransactionManager {
	return &InMemoryTransactionManager{}
}

func (tm *InMemoryTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if err := fn(ctx); err != nil {
		// InMemory ではロールバックの概念がないが、
		// 本番では tx.Rollback() が実行される。
		return fmt.Errorf("transaction rolled back: %w", err)
	}
	// 本番では tx.Commit() が実行される。
	return nil
}
