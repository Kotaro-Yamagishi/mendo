package menu

import (
	"context"
	"fmt"

	"mendo/internal/domain/menu"
)

// SoldOutMenuUsecase はメニューを品切れにするユースケース。
// Menu 集約の SoldOut() コマンドを使う。
type SoldOutMenuUsecase struct {
	menuReader menu.Reader
	menuWriter menu.Writer
}

func NewSoldOutMenuUsecase(mr menu.Reader, mw menu.Writer) *SoldOutMenuUsecase {
	return &SoldOutMenuUsecase{menuReader: mr, menuWriter: mw}
}

func (uc *SoldOutMenuUsecase) Execute(ctx context.Context, menuID menu.MenuID) error {
	m, err := uc.menuReader.FindByID(ctx, menuID)
	if err != nil {
		return fmt.Errorf("failed to find menu: %w", err)
	}

	m.SoldOut()

	if err := uc.menuWriter.Save(ctx, m); err != nil {
		return fmt.Errorf("failed to save menu: %w", err)
	}
	return nil
}
