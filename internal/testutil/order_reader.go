package testutil

import "context"

// StubOrderReader は order.Reader の Stub。
type StubOrderReader struct {
	PendingCount int
	CountErr     error
}

func (s *StubOrderReader) CountPending(_ context.Context) (int, error) {
	if s.CountErr != nil {
		return 0, s.CountErr
	}
	return s.PendingCount, nil
}
