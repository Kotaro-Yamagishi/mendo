package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"mendo/internal/apperrors"
	"mendo/internal/domain"
	"mendo/internal/infrastructure/datasource"
)

// OutboxRepository は datasource を使った Outbox の永続化実装。
// domain.Outbox を実装する。
type OutboxRepository struct {
	ds datasource.OutboxDataSource
}

func NewOutboxRepository(ds datasource.OutboxDataSource) *OutboxRepository {
	return &OutboxRepository{ds: ds}
}

// Store はイベント列を OutboxRow に変換して永続化する。
// EventStore.Save() と同じトランザクション内で呼ぶ。
func (r *OutboxRepository) Store(ctx context.Context, events []domain.Event) error {
	rows := make([]datasource.OutboxRow, 0, len(events))
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return apperrors.Infrastructure("イベントの変換に失敗", err)
		}
		rows = append(rows, datasource.OutboxRow{
			ID:          uuid.New().String(),
			EventType:   event.GetEventType(),
			AggregateID: event.GetAggregateID(),
			Payload:     string(payload),
			Delivered:   false,
			CreatedAt:   time.Now().UTC(),
		})
	}
	if err := r.ds.InsertOutboxRows(ctx, rows); err != nil {
		return apperrors.Infrastructure("アウトボックスへの保存に失敗", err)
	}
	return nil
}

// Fetch は未配信のイベントを limit 件取得する。
func (r *OutboxRepository) Fetch(ctx context.Context, limit int) ([]domain.Event, error) {
	rows, err := r.ds.FindUndeliveredOutboxRows(ctx, limit)
	if err != nil {
		return nil, apperrors.Infrastructure("アウトボックスの取得に失敗", err)
	}

	events := make([]domain.Event, 0, len(rows))
	for _, row := range rows {
		event, err := unmarshalOutboxEvent(row)
		if err != nil {
			return nil, apperrors.Infrastructure("アウトボックスイベントの変換に失敗", err)
		}
		events = append(events, event)
	}
	return events, nil
}

// MarkDelivered は配信済みのイベントをマークする。
func (r *OutboxRepository) MarkDelivered(ctx context.Context, eventIDs []string) error {
	if err := r.ds.MarkOutboxRowsDelivered(ctx, eventIDs); err != nil {
		return apperrors.Infrastructure("アウトボックスの配信済みマークに失敗", err)
	}
	return nil
}

// unmarshalOutboxEvent は OutboxRow から domain.Event を復元する。
// eventstore_repo.go の unmarshalEvent と同じロジックを使う。
func unmarshalOutboxEvent(row datasource.OutboxRow) (domain.Event, error) {
	eventRow := datasource.EventRow{
		EventType:   row.EventType,
		AggregateID: row.AggregateID,
		Payload:     row.Payload,
	}
	return unmarshalEvent(eventRow)
}
