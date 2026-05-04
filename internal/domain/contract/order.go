// Package contract は Bounded Context 間の公開イベントの型を定義する。
// BC 間のデータ交換の「契約」。
//
// ルール:
//   - 公開イベントの構造体と定数のみ
//   - IF やビジネスロジックは置かない
//   - 内部イベントや集約の型は置かない
//   - マイクロサービス化した時にこのパッケージを別リポジトリに切り出す
package contract

// --- Order BC の公開イベント ---

const (
	PublicEventTypeOrderCreated   = "public.order.created"
	PublicEventTypeOrderConfirmed = "public.order.confirmed"
	PublicEventTypeOrderCanceled  = "public.order.canceled"
)

// OrderConfirmedPublicItem は公開イベントに載せる注文明細。
// Order の内部構造（Item struct）とは別物。厨房が必要な情報だけ。
type OrderConfirmedPublicItem struct {
	MenuName string   `json:"menu_name"`
	Toppings []string `json:"toppings"`
	Hardness string   `json:"hardness"`
}

// OrderConfirmedPublic は他の BC に渡す用の注文確定イベント。
// Order の Items 構造体や DomainEvent メタデータは含まない。
type OrderConfirmedPublic struct {
	OrderID string                     `json:"order_id"`
	SeatNo  int                        `json:"seat_no"`
	Items   []OrderConfirmedPublicItem `json:"items"`
}

func (e OrderConfirmedPublic) GetEventType() string    { return PublicEventTypeOrderConfirmed }
func (e OrderConfirmedPublic) GetAggregateID() string  { return e.OrderID }
func (e OrderConfirmedPublic) GetCorrelationID() string { return "" }

type OrderCanceledPublic struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

func (e OrderCanceledPublic) GetEventType() string    { return PublicEventTypeOrderCanceled }
func (e OrderCanceledPublic) GetAggregateID() string  { return e.OrderID }
func (e OrderCanceledPublic) GetCorrelationID() string { return "" }

type OrderCreatedPublic struct {
	OrderID string `json:"order_id"`
	SeatNo  int    `json:"seat_no"`
}

func (e OrderCreatedPublic) GetEventType() string    { return PublicEventTypeOrderCreated }
func (e OrderCreatedPublic) GetAggregateID() string  { return e.OrderID }
func (e OrderCreatedPublic) GetCorrelationID() string { return "" }
