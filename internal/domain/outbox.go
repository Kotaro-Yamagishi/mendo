package domain

import "context"

// Outbox はイベントの確実な配信を保証する仕組み。
// DB トランザクション内で events と outbox を同時に保存し、
// 別プロセスが outbox を読んで EventBus に配信する。
//
// 第9章: 送信箱（Outbox Pattern）
//
// なぜ必要か:
//
//	eventStore.Save() と eventPublisher.Publish() は別のトランザクション。
//	Save は成功したのに Publish が失敗するとイベントが消える。
//	Outbox は同じ DB トランザクションで保存するので確実に届く。
type Outbox interface {
	// Store はイベントを outbox に保存する。
	// EventStore.Save() と同じトランザクション内で呼ぶ。
	Store(ctx context.Context, events []Event) error

	// Fetch は未配信のイベントを取得する。
	// 別プロセス（リレーサービス）が定期的に呼ぶ。
	Fetch(ctx context.Context, limit int) ([]Event, error)

	// MarkDelivered は配信済みのイベントをマークする。
	MarkDelivered(ctx context.Context, eventIDs []string) error
}
