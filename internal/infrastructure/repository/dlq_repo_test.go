package repository_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"mendo/internal/domain"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/datasource"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Store
// =============================================================================

func Test_DLQRepo_Store_正常系(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubDLQDataSource{}
	repo := repository.NewDLQRepository(ds)

	event := order.NewOrderCreated("order-1", 3, "corr-1")
	letter := &domain.DeadLetter{
		ID:          "dlq-1",
		Event:       event,
		Error:       "handler panicked",
		FailCount:   3,
		HandlerName: "KitchenSubscriber",
		LastFailAt:  time.Now(),
	}

	err := repo.Store(context.Background(), letter)

	require.NoError(t, err)
	require.NotNil(t, ds.InsertedRow)
	assert.Equal(t, "dlq-1", ds.InsertedRow.ID)
	assert.Equal(t, order.EventTypeOrderCreated, ds.InsertedRow.EventType)
	assert.Equal(t, "handler panicked", ds.InsertedRow.Error)
	assert.Equal(t, 3, ds.InsertedRow.FailCount)
	assert.Equal(t, "KitchenSubscriber", ds.InsertedRow.HandlerName)
	assert.NotEmpty(t, ds.InsertedRow.Payload)
}

func Test_DLQRepo_Store_InsertエラーがPropagateされる(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubDLQDataSource{InsertErr: errors.New("insert failed")}
	repo := repository.NewDLQRepository(ds)

	event := order.NewOrderCreated("order-1", 3, "")
	letter := &domain.DeadLetter{ID: "dlq-1", Event: event}

	err := repo.Store(context.Background(), letter)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DLQへの保存に失敗")
}

// =============================================================================
// List
// =============================================================================

func Test_DLQRepo_List_正常系(t *testing.T) {
	t.Parallel()

	// Payload に DomainEvent を JSON でセットする
	base := domain.NewDomainEvent("order-1", "Order", order.EventTypeOrderCreated, "")
	payload, _ := json.Marshal(base)

	tests := []struct {
		name    string
		rows    []datasource.DeadLetterRow
		wantLen int
	}{
		{
			name: "複数件",
			rows: []datasource.DeadLetterRow{
				{
					ID:          "dlq-1",
					EventType:   order.EventTypeOrderCreated,
					Payload:     string(payload),
					Error:       "timeout",
					FailCount:   1,
					HandlerName: "SomeHandler",
					LastFailAt:  time.Now(),
				},
			},
			wantLen: 1,
		},
		{
			name:    "空",
			rows:    nil,
			wantLen: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ds := &testutil.StubDLQDataSource{Rows: tc.rows}
			repo := repository.NewDLQRepository(ds)

			letters, err := repo.List(context.Background())

			require.NoError(t, err)
			assert.Len(t, letters, tc.wantLen)
		})
	}
}

func Test_DLQRepo_List_DataSourceエラー(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubDLQDataSource{FindAllErr: errors.New("db error")}
	repo := repository.NewDLQRepository(ds)

	_, err := repo.List(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DLQ一覧の取得に失敗")
}

// =============================================================================
// FindByID
// =============================================================================

func Test_DLQRepo_FindByID_正常系(t *testing.T) {
	t.Parallel()

	base := domain.NewDomainEvent("order-1", "Order", order.EventTypeOrderCreated, "")
	payload, _ := json.Marshal(base)

	ds := &testutil.StubDLQDataSource{
		SingleRow: &datasource.DeadLetterRow{
			ID:          "dlq-1",
			EventType:   order.EventTypeOrderCreated,
			Payload:     string(payload),
			FailCount:   2,
			HandlerName: "Handler",
		},
	}
	repo := repository.NewDLQRepository(ds)

	letter, err := repo.FindByID(context.Background(), "dlq-1")

	require.NoError(t, err)
	require.NotNil(t, letter)
	assert.Equal(t, "dlq-1", letter.ID)
	assert.Equal(t, 2, letter.FailCount)
}

func Test_DLQRepo_FindByID_見つからない(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubDLQDataSource{SingleRow: nil}
	repo := repository.NewDLQRepository(ds)

	_, err := repo.FindByID(context.Background(), "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =============================================================================
// Remove
// =============================================================================

func Test_DLQRepo_Remove_正常系(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubDLQDataSource{}
	repo := repository.NewDLQRepository(ds)

	err := repo.Remove(context.Background(), "dlq-1")

	require.NoError(t, err)
	assert.Equal(t, "dlq-1", ds.DeletedID)
}

func Test_DLQRepo_Remove_エラー(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubDLQDataSource{DeleteErr: errors.New("delete failed")}
	repo := repository.NewDLQRepository(ds)

	err := repo.Remove(context.Background(), "dlq-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DLQの削除に失敗")
}
