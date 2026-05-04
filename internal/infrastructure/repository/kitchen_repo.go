package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/datasource"
)

// KitchenRepository は datasource を使った Kitchen の永続化実装。
// kitchen.Reader と kitchen.Writer を実装する。
type KitchenRepository struct {
	ds datasource.KitchenDataSource
}

func NewKitchenRepository(ds datasource.KitchenDataSource) *KitchenRepository {
	return &KitchenRepository{ds: ds}
}

// FindByID は KitchenRow と CookingTaskRow[] を取得し、ドメインモデルを復元する。
func (r *KitchenRepository) FindByID(ctx context.Context, id kitchen.KitchenID) (*kitchen.Kitchen, error) {
	row, err := r.ds.FindKitchenByID(ctx, id.String())
	if err != nil {
		return nil, fmt.Errorf("KitchenRepository.FindByID: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("kitchen not found: %s", id)
	}

	taskRows, err := r.ds.FindCookingTasksByKitchenID(ctx, id.String())
	if err != nil {
		return nil, fmt.Errorf("KitchenRepository.FindByID tasks: %w", err)
	}

	tasks := make([]kitchen.CookingTask, 0, len(taskRows))
	for _, tr := range taskRows {
		var instrDTOs []datasource.CookingInstructionDTO
		if err := json.Unmarshal([]byte(tr.Instructions), &instrDTOs); err != nil {
			return nil, fmt.Errorf("KitchenRepository: unmarshal instructions: %w", err)
		}
		instructions := make([]kitchen.CookingInstruction, len(instrDTOs))
		for i, d := range instrDTOs {
			instructions[i] = kitchen.CookingInstruction{
				MenuName: d.MenuName,
				Toppings: d.Toppings,
				Hardness: d.Hardness,
			}
		}
		task := kitchen.ReconstructCookingTask(
			kitchen.TaskID(tr.TaskID),
			order.OrderID(tr.OrderID),
			instructions,
			kitchen.TaskStatus(taskStatusFromString(tr.Status)),
			tr.StartedAt,
		)
		tasks = append(tasks, task)
	}

	return kitchen.ReconstructKitchen(kitchen.KitchenID(row.KitchenID), tasks), nil
}

// Save は Kitchen を永続化する。
// DomainEvents に CookingCompleted があれば対応タスクのステータスを更新する。
func (r *KitchenRepository) Save(ctx context.Context, k *kitchen.Kitchen) error {
	kitchenRow := &datasource.KitchenRow{
		KitchenID:   k.ID().String(),
		MaxCapacity: kitchen.MaxConcurrentTasks,
	}
	if err := r.ds.UpsertKitchen(ctx, kitchenRow); err != nil {
		return fmt.Errorf("KitchenRepository.Save UpsertKitchen: %w", err)
	}

	// DomainEvents から CookingCompleted を検出してタスクステータスを更新
	for _, event := range k.DomainEvents() {
		if e, ok := event.(kitchen.CookingCompleted); ok {
			if err := r.ds.UpdateCookingTaskStatus(ctx, k.ID().String(), string(e.OrderID), "completed"); err != nil {
				return fmt.Errorf("KitchenRepository.Save UpdateCookingTaskStatus: %w", err)
			}
		}
	}

	return nil
}

func taskStatusFromString(s string) int {
	switch s {
	case "cooking":
		return int(kitchen.TaskCooking)
	case "completed":
		return int(kitchen.TaskCompleted)
	default:
		return int(kitchen.TaskPending)
	}
}
