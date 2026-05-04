package datasource

import "context"

// KitchenDataSource は kc_kitchens / kc_cooking_tasks テーブルへのアクセス IF。
// 引数・戻り値はすべてプリミティブ型または DTO。ドメイン型を使わない。
type KitchenDataSource interface {
	// FindKitchenByID は kitchen_id を指定して KitchenRow を返す。
	// 見つからない場合は nil, nil を返す。
	FindKitchenByID(ctx context.Context, id string) (*KitchenRow, error)

	// FindCookingTasksByKitchenID は kitchen_id に紐づく全タスクを返す。
	FindCookingTasksByKitchenID(ctx context.Context, kitchenID string) ([]CookingTaskRow, error)

	// UpsertKitchen は Kitchen を INSERT OR UPDATE する。
	UpsertKitchen(ctx context.Context, row *KitchenRow) error

	// InsertCookingTask は新しい CookingTask を INSERT する。
	InsertCookingTask(ctx context.Context, row *CookingTaskRow) error

	// UpdateCookingTaskStatus は指定した kitchen_id + order_id のタスクのステータスを更新する。
	UpdateCookingTaskStatus(ctx context.Context, kitchenID, orderID string, status string) error
}
