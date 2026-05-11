package testutil

import (
	"context"
	"fmt"

	"mendo/internal/domain"
)

// StubDeadLetterQueue は domain.DeadLetterQueue の Stub + Spy。
type StubDeadLetterQueue struct {
	Letters   []domain.DeadLetter
	SingleLetter *domain.DeadLetter
	StoreErr  error
	ListErr   error
	FindErr   error
	RemoveErr error

	StoredLetters []domain.DeadLetter
	RemovedIDs    []string
}

func (s *StubDeadLetterQueue) Store(_ context.Context, letter *domain.DeadLetter) error {
	if s.StoreErr != nil {
		return s.StoreErr
	}
	s.StoredLetters = append(s.StoredLetters, *letter)
	return nil
}

func (s *StubDeadLetterQueue) List(_ context.Context) ([]domain.DeadLetter, error) {
	if s.ListErr != nil {
		return nil, s.ListErr
	}
	return s.Letters, nil
}

func (s *StubDeadLetterQueue) Remove(_ context.Context, id string) error {
	if s.RemoveErr != nil {
		return s.RemoveErr
	}
	s.RemovedIDs = append(s.RemovedIDs, id)
	return nil
}

func (s *StubDeadLetterQueue) FindByID(_ context.Context, id string) (*domain.DeadLetter, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	if s.SingleLetter != nil {
		return s.SingleLetter, nil
	}
	for i := range s.Letters {
		if s.Letters[i].ID == id {
			return &s.Letters[i], nil
		}
	}
	return nil, fmt.Errorf("dead letter not found: %s", id)
}
