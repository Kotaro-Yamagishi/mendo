package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DomainEvent は全ドメインイベントが埋め込む基盤構造体。
// イベントの追跡・べき等性確認・因果関係の把握に必要なメタデータを持つ。
type DomainEvent struct {
	// EventID はイベントの一意識別子。重複配信時のべき等性確認に使う。
	EventID string `json:"event_id"`

	// AggregateID はイベントを発行した集約のID。
	AggregateID string `json:"aggregate_id"`

	// AggregateType はイベントを発行した集約の種別（"Order", "Kitchen" 等）。
	// 各集約で定数として定義する（例: order.AggregateTypeOrder）。
	AggregateType string `json:"aggregate_type"`

	// EventType はイベントの種別（"order.confirmed", "cooking.completed" 等）。
	// 購読者へのディスパッチキーになる。各集約で定数として定義する。
	EventType string `json:"event_type"`

	// OccurredAt はイベントが発生した時刻。
	OccurredAt time.Time `json:"occurred_at"`

	// CorrelationID はリクエスト全体を追跡するID。
	// 1つのHTTPリクエストから複数のイベントが発生しても、全て同じCorrelationIDを持つ。
	CorrelationID string `json:"correlation_id"`

	// CausationID はこのイベントの直接の原因となったイベントのID。
	// イベント連鎖: OrderConfirmed(causation) → CookingStarted(this)
	CausationID string `json:"causation_id"`
}

// NewDomainEvent は基盤イベントを生成する。
func NewDomainEvent(aggregateID, aggregateType, eventType, correlationID string) DomainEvent {
	return DomainEvent{
		EventID:       uuid.New().String(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventType:     eventType,
		OccurredAt:    time.Now().UTC(),
		CorrelationID: correlationID,
	}
}

// WithCausation は因果関係を設定して返す。
func (e DomainEvent) WithCausation(causationID string) DomainEvent {
	e.CausationID = causationID
	return e
}

// Event はドメインイベントの共通インターフェース。
type Event interface {
	GetEventType() string
	GetAggregateID() string
	GetCorrelationID() string
}

// EventPublisher はドメインイベントを配信するインターフェース。
// domain 層で定義し、infrastructure 層で実装する。
type EventPublisher interface {
	Publish(ctx context.Context, events ...Event) error
}
