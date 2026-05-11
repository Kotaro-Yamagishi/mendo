package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"mendo/internal/domain"
	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/datasource"
)

// EventStoreRepository は datasource を使ったイベントストアの永続化実装。
// domain.EventStore を実装する。
type EventStoreRepository struct {
	ds datasource.EventStoreDataSource
}

func NewEventStoreRepository(ds datasource.EventStoreDataSource) *EventStoreRepository {
	return &EventStoreRepository{ds: ds}
}

// Save はイベント列を EventRow に変換して永続化する。
func (r *EventStoreRepository) Save(ctx context.Context, events []domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	// 既存イベント数を取得して version のオフセットにする。
	// 2回目以降の Save で version が 0 からリセットされる問題を防ぐ。
	aggregateID := events[0].GetAggregateID()
	existing, err := r.ds.FindEventsByAggregateID(ctx, aggregateID)
	if err != nil {
		return fmt.Errorf("EventStoreRepository.Save load existing: %w", err)
	}
	versionOffset := len(existing)

	rows := make([]datasource.EventRow, 0, len(events))
	for i, event := range events {
		payload, err := marshalEvent(event)
		if err != nil {
			return fmt.Errorf("EventStoreRepository.Save marshal[%d]: %w", i, err)
		}
		rows = append(rows, datasource.EventRow{
			EventID:       uuid.New().String(),
			AggregateID:   event.GetAggregateID(),
			EventType:     event.GetEventType(),
			CorrelationID: event.GetCorrelationID(),
			Payload:       string(payload),
			Version:       versionOffset + i,
			CreatedAt:     time.Now().UTC(),
		})
	}
	if err := r.ds.InsertEvents(ctx, rows); err != nil {
		return fmt.Errorf("EventStoreRepository.Save InsertEvents: %w", err)
	}
	return nil
}

// Load は aggregate_id に紐づくイベント列を時系列順に取得する。
// event_type に応じてデシリアライズし、domain.Event スライスで返す。
func (r *EventStoreRepository) Load(ctx context.Context, aggregateID string) ([]domain.Event, error) {
	rows, err := r.ds.FindEventsByAggregateID(ctx, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("EventStoreRepository.Load: %w", err)
	}

	events := make([]domain.Event, 0, len(rows))
	for _, row := range rows {
		event, err := unmarshalEvent(row)
		if err != nil {
			return nil, fmt.Errorf("EventStoreRepository.Load unmarshal %s: %w", row.EventType, err)
		}
		events = append(events, event)
	}
	return events, nil
}

// marshalEvent はドメインイベントを JSON にシリアライズする。
func marshalEvent(event domain.Event) ([]byte, error) {
	return json.Marshal(event)
}

// unmarshalEvent は EventRow の event_type に応じてドメインイベントを復元する。
// 学習用のため、未知のイベントは DomainEvent ラッパーで返す。
func unmarshalEvent(row datasource.EventRow) (domain.Event, error) {
	switch row.EventType {
	case order.EventTypeOrderCreated:
		var e order.OrderCreated
		if err := json.Unmarshal([]byte(row.Payload), &e); err != nil {
			return nil, err
		}
		return e, nil
	case order.EventTypeItemAdded:
		var e order.ItemAdded
		if err := json.Unmarshal([]byte(row.Payload), &e); err != nil {
			return nil, err
		}
		return e, nil
	case order.EventTypeOrderConfirmed:
		var e order.OrderConfirmed
		if err := json.Unmarshal([]byte(row.Payload), &e); err != nil {
			return nil, err
		}
		return e, nil
	case order.EventTypeOrderCanceled:
		var e order.OrderCancelled
		if err := json.Unmarshal([]byte(row.Payload), &e); err != nil {
			return nil, err
		}
		return e, nil
	case kitchen.EventTypeCookingCompleted:
		var e kitchen.CookingCompleted
		if err := json.Unmarshal([]byte(row.Payload), &e); err != nil {
			return nil, err
		}
		return e, nil
	case kitchen.EventTypeCookingRejected:
		var e kitchen.CookingRejected
		if err := json.Unmarshal([]byte(row.Payload), &e); err != nil {
			return nil, err
		}
		return e, nil
	default:
		// 未知のイベントは DomainEvent ラッパーで返す（学習用）
		var base domain.DomainEvent
		if err := json.Unmarshal([]byte(row.Payload), &base); err != nil {
			return nil, err
		}
		return &unknownEvent{DomainEvent: base}, nil
	}
}

// unknownEvent は未知のイベント種別を保持するラッパー。
type unknownEvent struct {
	domain.DomainEvent
}

func (e *unknownEvent) GetEventType() string     { return e.EventType }
func (e *unknownEvent) GetAggregateID() string   { return e.AggregateID }
func (e *unknownEvent) GetCorrelationID() string { return e.CorrelationID }
