package repository_test

import (
	"context"
	"errors"
	"testing"

	"mendo/internal/domain"
	"mendo/internal/domain/menu"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/datasource"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Save — ドメインイベント → EventRow への変換を検証
// =============================================================================

func Test_EventStoreRepo_Save_正常系(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubEventStoreDataSource{}
	repo := repository.NewEventStoreRepository(ds)

	events := []domain.Event{
		order.NewOrderCreated("order-1", 3, "corr-1"),
		order.NewItemAdded("order-1", menu.MenuID("menu-1"), []string{"ネギ"}, "硬め", "corr-1"),
	}

	err := repo.Save(context.Background(), events)

	require.NoError(t, err)
	require.Len(t, ds.Inserted, 2, "2つの EventRow が InsertEvents に渡される")

	// EventRow のマッピング検証
	assert.NotEmpty(t, ds.Inserted[0].EventID, "EventID が UUID で設定されている")
	assert.Equal(t, "order-1", ds.Inserted[0].AggregateID)
	assert.Equal(t, order.EventTypeOrderCreated, ds.Inserted[0].EventType)
	assert.Equal(t, "corr-1", ds.Inserted[0].CorrelationID)
	assert.Equal(t, 0, ds.Inserted[0].Version)
	assert.NotEmpty(t, ds.Inserted[0].Payload, "Payload が JSON シリアライズされている")

	assert.Equal(t, 1, ds.Inserted[1].Version)
	assert.Equal(t, order.EventTypeItemAdded, ds.Inserted[1].EventType)
}

func Test_EventStoreRepo_Save_バリエーション(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		existingRows    []datasource.EventRow
		inputEvents     []domain.Event
		wantInsertLen   int
		wantFirstVer    int // wantInsertLen > 0 のときのみ検証
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "versionオフセット: 既存2件の場合は version 2 から始まる",
			existingRows: []datasource.EventRow{
				{Version: 0}, {Version: 1},
			},
			inputEvents: []domain.Event{
				order.NewOrderConfirmed("order-1", []order.ConfirmedItem{
					{MenuID: "menu-1", Toppings: nil, Hardness: "普通"},
				}, 3, ""),
			},
			wantInsertLen: 1,
			wantFirstVer:  2,
		},
		{
			name:          "空スライスでは InsertEvents が呼ばれない",
			inputEvents:   []domain.Event{},
			wantInsertLen: 0,
		},
		{
			name: "InsertError が伝播する",
			inputEvents: []domain.Event{
				order.NewOrderCreated("order-1", 1, ""),
			},
			wantErr:         true,
			wantErrContains: "InsertEvents",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var insertErr error
			if tc.wantErr {
				insertErr = errors.New("db error")
			}
			ds := &testutil.StubEventStoreDataSource{
				Events:    tc.existingRows,
				InsertErr: insertErr,
			}
			repo := repository.NewEventStoreRepository(ds)

			err := repo.Save(context.Background(), tc.inputEvents)

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrContains)
				return
			}
			require.NoError(t, err)
			require.Len(t, ds.Inserted, tc.wantInsertLen)
			if tc.wantInsertLen > 0 {
				assert.Equal(t, tc.wantFirstVer, ds.Inserted[0].Version)
			}
		})
	}
}

// =============================================================================
// Load — EventRow → ドメインイベントへの変換を検証
// =============================================================================

func Test_EventStoreRepo_Load_正常系(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubEventStoreDataSource{
		Events: []datasource.EventRow{
			{
				EventID:     "evt-1",
				AggregateID: "order-1",
				EventType:   order.EventTypeOrderCreated,
				Payload:     `{"event_id":"evt-1","aggregate_id":"order-1","aggregate_type":"Order","event_type":"order.created","occurred_at":"0001-01-01T00:00:00Z","correlation_id":"","causation_id":"","seat_no":3}`,
				Version:     0,
			},
			{
				EventID:     "evt-2",
				AggregateID: "order-1",
				EventType:   order.EventTypeItemAdded,
				Payload:     `{"event_id":"evt-2","aggregate_id":"order-1","aggregate_type":"Order","event_type":"order.item_added","occurred_at":"0001-01-01T00:00:00Z","correlation_id":"","causation_id":"","menu_id":"menu-1","toppings":["ネギ"],"hardness":"硬め"}`,
				Version:     1,
			},
		},
	}
	repo := repository.NewEventStoreRepository(ds)

	events, err := repo.Load(context.Background(), "order-1")

	require.NoError(t, err)
	require.Len(t, events, 2)

	// 型の検証: event_type に応じて正しいドメインイベント型に変換される
	created, ok := events[0].(order.OrderCreated)
	require.True(t, ok, "order.created → OrderCreated 型")
	assert.Equal(t, "order-1", created.AggregateID)
	assert.Equal(t, 3, created.SeatNo)

	added, ok := events[1].(order.ItemAdded)
	require.True(t, ok, "order.item_added → ItemAdded 型")
	assert.Equal(t, menu.MenuID("menu-1"), added.MenuID)
}

func Test_EventStoreRepo_Load_バリエーション(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		stubEvents      []datasource.EventRow
		findErr         error
		wantLen         int
		wantErr         bool
		wantErrContains string
	}{
		{
			name:       "空: イベントなしの集約",
			stubEvents: nil,
			wantLen:    0,
		},
		{
			name:            "FindError が伝播する",
			findErr:         errors.New("db error"),
			wantErr:         true,
			wantErrContains: "Load",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ds := &testutil.StubEventStoreDataSource{
				Events:  tc.stubEvents,
				FindErr: tc.findErr,
			}
			repo := repository.NewEventStoreRepository(ds)

			events, err := repo.Load(context.Background(), "order-1")

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrContains)
				return
			}
			require.NoError(t, err)
			assert.Len(t, events, tc.wantLen)
		})
	}
}
