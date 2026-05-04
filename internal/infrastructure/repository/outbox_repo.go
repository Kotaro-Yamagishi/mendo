package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
			return fmt.Errorf("OutboxRepository.Store marshal: %w", err)
		}
		rows = append(rows, datasource.OutboxRow{
			EventType:   event.GetEventType(),
			AggregateID: event.GetAggregateID(),
			Payload:     string(payload),
			Delivered:   false,
			CreatedAt:   time.Now().UTC(),
		})
	}
	if err := r.ds.InsertOutboxRows(ctx, rows); err != nil {
		return fmt.Errorf("OutboxRepository.Store InsertOutboxRows: %w", err)
	}
	return nil
}

// Fetch は未配信のイベントを limit 件取得する。
func (r *OutboxRepository) Fetch(ctx context.Context, limit int) ([]domain.Event, error) {
	rows, err := r.ds.FindUndeliveredOutboxRows(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("OutboxRepository.Fetch: %w", err)
	}

	events := make([]domain.Event, 0, len(rows))
	for _, row := range rows {
		event, err := unmarshalOutboxEvent(row)
		if err != nil {
			return nil, fmt.Errorf("OutboxRepository.Fetch unmarshal %s: %w", row.EventType, err)
		}
		events = append(events, event)
	}
	return events, nil
}

// MarkDelivered は配信済みのイベントをマークする。
func (r *OutboxRepository) MarkDelivered(ctx context.Context, eventIDs []string) error {
	if err := r.ds.MarkOutboxRowsDelivered(ctx, eventIDs); err != nil {
		return fmt.Errorf("OutboxRepository.MarkDelivered: %w", err)
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
