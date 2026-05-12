package repository

import (
	"context"
	"encoding/json"
	"time"

	"mendo/internal/apperrors"
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
		return nil, apperrors.Infrastructure("厨房の取得に失敗", err)
	}
	if row == nil {
		return nil, apperrors.NotFound("kitchen", string(id))
	}

	taskRows, err := r.ds.FindCookingTasksByKitchenID(ctx, id.String())
	if err != nil {
		return nil, apperrors.Infrastructure("調理タスクの取得に失敗", err)
	}

	tasks := make([]kitchen.CookingTask, 0, len(taskRows))
	for _, tr := range taskRows {
		var instrDTOs []datasource.CookingInstructionDTO
		if err := json.Unmarshal([]byte(tr.Instructions), &instrDTOs); err != nil {
			return nil, apperrors.Infrastructure("調理手順の変換に失敗", err)
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
// 全タスクを INSERT IGNORE で保存し、DomainEvents に CookingCompleted があれば対応タスクのステータスを更新する。
func (r *KitchenRepository) Save(ctx context.Context, k *kitchen.Kitchen) error {
	kitchenRow := &datasource.KitchenRow{
		KitchenID:   k.ID().String(),
		MaxCapacity: kitchen.MaxConcurrentTasks,
		CreatedAt:   time.Now().UTC(),
	}
	if err := r.ds.UpsertKitchen(ctx, kitchenRow); err != nil {
		return apperrors.Infrastructure("厨房の保存に失敗", err)
	}

	// 全タスクを INSERT IGNORE（既存行は無視）
	for _, task := range k.Tasks() {
		instrDTOs := make([]datasource.CookingInstructionDTO, len(task.Instructions()))
		for i, instr := range task.Instructions() {
			instrDTOs[i] = datasource.CookingInstructionDTO{
				MenuName: instr.MenuName,
				Toppings: instr.Toppings,
				Hardness: instr.Hardness,
			}
		}
		instrJSON, err := json.Marshal(instrDTOs)
		if err != nil {
			return apperrors.Infrastructure("調理手順の変換に失敗", err)
		}
		taskRow := &datasource.CookingTaskRow{
			TaskID:       string(task.ID()),
			KitchenID:    k.ID().String(),
			OrderID:      string(task.OrderID()),
			Status:       taskStatusToString(kitchen.TaskStatus(task.Status())),
			Instructions: string(instrJSON),
			StartedAt:    task.CreatedAt().UTC(),
		}
		if err := r.ds.InsertCookingTask(ctx, taskRow); err != nil {
			return apperrors.Infrastructure("調理タスクの保存に失敗", err)
		}
	}

	// DomainEvents から CookingCompleted を検出してタスクステータスを更新
	for _, event := range k.DomainEvents() {
		if e, ok := event.(kitchen.CookingCompleted); ok {
			if err := r.ds.UpdateCookingTaskStatus(ctx, k.ID().String(), string(e.OrderID), "completed"); err != nil {
				return apperrors.Infrastructure("調理タスクのステータス更新に失敗", err)
			}
		}
	}

	return nil
}

func taskStatusToString(s kitchen.TaskStatus) string {
	switch s {
	case kitchen.TaskCooking:
		return "cooking"
	case kitchen.TaskCompleted:
		return "completed"
	default:
		return "pending"
	}
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
