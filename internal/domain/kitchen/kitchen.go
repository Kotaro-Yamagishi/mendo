package kitchen

import (
	"mendo/internal/domain"
	"mendo/internal/apperrors"
	"mendo/internal/domain/order"
)

// Kitchen は厨房集約のルート。調理タスクのライフサイクルを管理する。
type Kitchen struct {
	id           KitchenID
	tasks        []CookingTask
	domainEvents []domain.Event
}

func NewKitchen(id KitchenID) *Kitchen {
	return &Kitchen{id: id}
}

// AddCookingTask は調理タスクを追加する。OrderConfirmed イベントの購読者から呼ばれる。
func (k *Kitchen) AddCookingTask(orderID order.OrderID, instructions []CookingInstruction) error {
	if len(instructions) == 0 {
		return apperrors.Domain(ErrCodeEmptyInstructions, "調理指示は1つ以上必要です")
	}
	// 業務ルール: 同時調理数の上限チェック
	activeTasks := 0
	for _, t := range k.tasks {
		if t.status != TaskCompleted {
			activeTasks++
		}
	}
	if activeTasks >= MaxConcurrentTasks {
		k.domainEvents = append(k.domainEvents, NewCookingRejected(k.id, string(orderID), "厨房がフル稼働中です", ""))
		return apperrors.Domain(ErrCodeKitchenFull, "厨房がフル稼働中です。しばらくお待ちください")
	}

	task := newCookingTask(orderID, instructions)
	k.tasks = append(k.tasks, task)
	return nil
}

// CompleteCookingTask は調理を完了にする。
func (k *Kitchen) CompleteCookingTask(orderID order.OrderID) error {
	for i := range k.tasks {
		if k.tasks[i].orderID == orderID {
			// 業務ルール: すでに完了してないかチェック
			if k.tasks[i].status == TaskCompleted {
				return apperrors.Domain(ErrCodeAlreadyCooked, "すでに調理完了です")
			}

			// 状態変更
			k.tasks[i].status = TaskCompleted

			// ドメインイベント発行
			k.domainEvents = append(k.domainEvents, NewCookingCompleted(k.id, orderID, ""))
			return nil
		}
	}
	return apperrors.Domain(ErrCodeTaskNotFound, "調理タスクが見つかりません")
}

// CookingCapacity は現在の調理可能数を返す。
func (k *Kitchen) CookingCapacity() int {
	activeTasks := 0
	for _, t := range k.tasks {
		if t.status != TaskCompleted {
			activeTasks++
		}
	}
	return MaxConcurrentTasks - activeTasks
}

func (k *Kitchen) ID() KitchenID               { return k.id }
func (k *Kitchen) DomainEvents() []domain.Event { return k.domainEvents }
func (k *Kitchen) Tasks() []CookingTask         { return k.tasks }

// ReconstructKitchen は DB から読み込んだタスク一覧を使って Kitchen を復元する。
// ドメインイベントは発行しない（永続化済みの状態を読み戻すだけ）。
// infrastructure 層のリポジトリ実装専用。
func ReconstructKitchen(id KitchenID, tasks []CookingTask) *Kitchen {
	return &Kitchen{
		id:    id,
		tasks: tasks,
	}
}
