package repository

import (
	"context"
	"encoding/json"

	"mendo/internal/apperrors"
	"mendo/internal/domain"
	"mendo/internal/infrastructure/datasource"
)

// DLQRepository は datasource を使った DeadLetterQueue の永続化実装。
// domain.DeadLetterQueue を実装する。
type DLQRepository struct {
	ds datasource.DLQDataSource
}

func NewDLQRepository(ds datasource.DLQDataSource) *DLQRepository {
	return &DLQRepository{ds: ds}
}

// Store は DeadLetter を DeadLetterRow に変換して永続化する。
func (r *DLQRepository) Store(ctx context.Context, letter *domain.DeadLetter) error {
	payload, err := json.Marshal(letter.Event)
	if err != nil {
		return apperrors.Infrastructure("DLQイベントの変換に失敗", err)
	}
	row := &datasource.DeadLetterRow{
		ID:          letter.ID,
		EventType:   letter.Event.GetEventType(),
		Payload:     string(payload),
		Error:       letter.Error,
		FailCount:   letter.FailCount,
		HandlerName: letter.HandlerName,
		LastFailAt:  letter.LastFailAt,
	}
	if err := r.ds.InsertDeadLetterRow(ctx, row); err != nil {
		return apperrors.Infrastructure("DLQへの保存に失敗", err)
	}
	return nil
}

// List は全 DeadLetter を返す。
func (r *DLQRepository) List(ctx context.Context) ([]domain.DeadLetter, error) {
	rows, err := r.ds.FindAllDeadLetterRows(ctx)
	if err != nil {
		return nil, apperrors.Infrastructure("DLQ一覧の取得に失敗", err)
	}
	letters := make([]domain.DeadLetter, 0, len(rows))
	for _, row := range rows {
		letter, err := rowToDeadLetter(&row)
		if err != nil {
			return nil, err
		}
		letters = append(letters, *letter)
	}
	return letters, nil
}

// Remove は指定した ID の DeadLetter を削除する。
func (r *DLQRepository) Remove(ctx context.Context, id string) error {
	if err := r.ds.DeleteDeadLetterRow(ctx, id); err != nil {
		return apperrors.Infrastructure("DLQの削除に失敗", err)
	}
	return nil
}

// FindByID は指定した ID の DeadLetter を返す。
func (r *DLQRepository) FindByID(ctx context.Context, id string) (*domain.DeadLetter, error) {
	row, err := r.ds.FindDeadLetterRowByID(ctx, id)
	if err != nil {
		return nil, apperrors.Infrastructure("DLQの取得に失敗", err)
	}
	if row == nil {
		return nil, apperrors.NotFound("dead_letter", id)
	}
	return rowToDeadLetter(row)
}

func rowToDeadLetter(row *datasource.DeadLetterRow) (*domain.DeadLetter, error) {
	// Payload から DomainEvent を復元する（学習用のため DomainEvent ラッパーを使う）
	var base domain.DomainEvent
	if err := json.Unmarshal([]byte(row.Payload), &base); err != nil {
		return nil, apperrors.Infrastructure("DLQイベントの変換に失敗", err)
	}
	event := &unknownEvent{DomainEvent: base}
	return &domain.DeadLetter{
		ID:          row.ID,
		Event:       event,
		Error:       row.Error,
		FailCount:   row.FailCount,
		HandlerName: row.HandlerName,
		LastFailAt:  row.LastFailAt,
	}, nil
}
