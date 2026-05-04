package datasource

import "time"

// ============================================================
// Kitchen コンテキスト
// ============================================================

// KitchenRow は kc_kitchens テーブルの1行に対応する DTO。
type KitchenRow struct {
	KitchenID   string
	MaxCapacity int
	CreatedAt   time.Time
}

// CookingTaskRow は kc_cooking_tasks テーブルの1行に対応する DTO。
type CookingTaskRow struct {
	TaskID      string
	KitchenID   string
	OrderID     string
	Status      string
	Instructions string // JSON 文字列（[]CookingInstructionDTO のシリアライズ）
	StartedAt   time.Time
	CompletedAt *time.Time
}

// CookingInstructionDTO は CookingTask の instructions JSON 要素に対応する DTO。
type CookingInstructionDTO struct {
	MenuName string   `json:"menu_name"`
	Toppings []string `json:"toppings"`
	Hardness string   `json:"hardness"`
}

// ============================================================
// Menu コンテキスト
// ============================================================

// MenuRow は oc_menus テーブルの1行に対応する DTO。
type MenuRow struct {
	MenuID    string
	Name      string
	Price     int
	Available bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ============================================================
// SpecialOrder コンテキスト
// ============================================================

// SpecialOrderRow は sc_special_orders テーブルの1行に対応する DTO。
type SpecialOrderRow struct {
	ID            string
	OrderID       string
	MenuName      string
	Status        int
	RejectReason  string
	SuggestedMenu string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ============================================================
// EventStore（共通）
// ============================================================

// EventRow は events テーブルの1行に対応する DTO。
type EventRow struct {
	EventID       string
	AggregateID   string
	AggregateType string
	EventType     string
	Version       int
	Payload       string // JSON 文字列
	CorrelationID string
	CausationID   string
	CreatedAt     time.Time
}

// ============================================================
// Outbox（共通）
// ============================================================

// OutboxRow は outbox テーブルの1行に対応する DTO。
type OutboxRow struct {
	ID          string
	EventType   string
	AggregateID string
	Payload     string // JSON 文字列
	Delivered   bool
	CreatedAt   time.Time
}

// ============================================================
// Dead Letter Queue（共通）
// ============================================================

// DeadLetterRow は dlq テーブルの1行に対応する DTO。
type DeadLetterRow struct {
	ID          string
	EventType   string
	Payload     string // JSON 文字列
	Error       string
	FailCount   int
	HandlerName string
	LastFailAt  time.Time
}

// ============================================================
// Order Projection（注文コンテキスト）
// ============================================================

// OrderProjectionRow は oc_order_projections テーブルの1行に対応する DTO。
type OrderProjectionRow struct {
	OrderID   string
	SeatNo    int
	Items     string // JSON 文字列
	Total     int
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ============================================================
// OrderBoard Projection（厨房コンテキスト横断）
// ============================================================

// OrderBoardRow は kc_order_board テーブルの1行に対応する DTO。
type OrderBoardRow struct {
	OrderID       string
	SeatNo        int
	OrderStatus   string
	CookingStatus string
	OrderedAt     *time.Time
	CookingAt     *time.Time
}

// ============================================================
// Staff（補完領域）
// ============================================================

// StaffRow は staffs テーブルの1行に対応する DTO。
type StaffRow struct {
	ID        string
	Name      string
	Phone     string
	ShiftType string
	CreatedAt time.Time
	UpdatedAt time.Time
}
