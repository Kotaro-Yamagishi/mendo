package kitchen

import (
	"mendo/internal/domain/order"
	"time"
)

// CookingInstruction は調理に必要な指示。公開イベントから受け取る情報。
type CookingInstruction struct {
	MenuName string
	Toppings []string
	Hardness string
}

// CookingTask は厨房集約の内部エンティティ。Kitchen 経由でのみ操作。
type CookingTask struct {
	id           TaskID
	orderID      order.OrderID
	instructions []CookingInstruction
	status       TaskStatus
	createdAt    time.Time
}

func newCookingTask(orderID order.OrderID, instructions []CookingInstruction) CookingTask {
	return CookingTask{
		id:           newTaskID(),
		orderID:      orderID,
		instructions: instructions,
		status:       TaskPending,
		createdAt:    time.Now(),
	}
}

func (t *CookingTask) ID() TaskID             { return t.id }
func (t *CookingTask) OrderID() order.OrderID { return t.orderID }
func (t *CookingTask) Status() TaskStatus     { return t.status }
func (t *CookingTask) Instructions() []CookingInstruction { return t.instructions }
func (t *CookingTask) CreatedAt() time.Time   { return t.createdAt }

// ReconstructCookingTask は DB から読み込んだ値を使って CookingTask を復元する。
// infrastructure 層のリポジトリ実装専用。
func ReconstructCookingTask(id TaskID, orderID order.OrderID, instructions []CookingInstruction, status TaskStatus, createdAt time.Time) CookingTask {
	return CookingTask{
		id:           id,
		orderID:      orderID,
		instructions: instructions,
		status:       status,
		createdAt:    createdAt,
	}
}
