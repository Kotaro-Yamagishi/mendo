package repository

import (
	"context"
	"fmt"

	"mendo/internal/domain/menu"
	"mendo/internal/infrastructure/datasource"
)

// MenuRepository は datasource を使った Menu の永続化実装。
// menu.Reader と menu.Writer を実装する。
type MenuRepository struct {
	ds datasource.MenuDataSource
}

func NewMenuRepository(ds datasource.MenuDataSource) *MenuRepository {
	return &MenuRepository{ds: ds}
}

// FindByID は MenuRow を取得してドメインモデルを復元する。
func (r *MenuRepository) FindByID(ctx context.Context, id menu.MenuID) (*menu.Menu, error) {
	row, err := r.ds.FindMenuByID(ctx, id.String())
	if err != nil {
		return nil, fmt.Errorf("MenuRepository.FindByID: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("menu not found: %s", id)
	}
	return rowToMenu(row)
}

// FindAll は全 MenuRow を取得してドメインモデルのスライスを返す。
func (r *MenuRepository) FindAll(ctx context.Context) ([]*menu.Menu, error) {
	rows, err := r.ds.FindAllMenus(ctx)
	if err != nil {
		return nil, fmt.Errorf("MenuRepository.FindAll: %w", err)
	}
	menus := make([]*menu.Menu, 0, len(rows))
	for i := range rows {
		m, err := rowToMenu(&rows[i])
		if err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}
	return menus, nil
}

// Save は Menu を MenuRow に変換して永続化する。
func (r *MenuRepository) Save(ctx context.Context, m *menu.Menu) error {
	row := &datasource.MenuRow{
		MenuID:    m.ID().String(),
		Name:      m.Name().String(),
		Price:     m.Price().Yen(),
		Available: m.IsAvailable(),
	}
	if err := r.ds.InsertMenu(ctx, row); err != nil {
		return fmt.Errorf("MenuRepository.Save: %w", err)
	}
	return nil
}

func rowToMenu(row *datasource.MenuRow) (*menu.Menu, error) {
	name, err := menu.NewMenuName(row.Name)
	if err != nil {
		return nil, fmt.Errorf("rowToMenu NewMenuName: %w", err)
	}
	price, err := menu.NewPrice(row.Price)
	if err != nil {
		return nil, fmt.Errorf("rowToMenu NewPrice: %w", err)
	}
	return menu.ReconstructMenu(menu.MenuID(row.MenuID), name, price, row.Available), nil
}
