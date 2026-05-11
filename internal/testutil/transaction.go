package testutil

import "context"

// StubTransactionManager はトランザクションの Stub。
// Do の中のコールバックをそのまま実行する（トランザクションなし）。
type StubTransactionManager struct {
	DoErr error
}

func (s *StubTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if s.DoErr != nil {
		return s.DoErr
	}
	return fn(ctx)
}
