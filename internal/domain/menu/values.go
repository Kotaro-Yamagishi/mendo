package menu

import (
	"fmt"

	"mendo/internal/apperrors"
)

// --- MenuID ---

type MenuID string

func (id MenuID) String() string { return string(id) }

// --- MenuName ---

type MenuName struct {
	value string
}

func NewMenuName(s string) (MenuName, error) {
	if s == "" {
		return MenuName{}, apperrors.Domain(ErrCodeInvalidMenuName, "メニュー名は空にできません")
	}
	return MenuName{value: s}, nil
}

func (n MenuName) String() string { return n.value }

// --- Price ---

type Price int

func NewPrice(yen int) (Price, error) {
	if yen < 0 {
		return 0, apperrors.Domain(ErrCodeInvalidPrice, fmt.Sprintf("価格は0以上: %d", yen))
	}
	return Price(yen), nil
}

func (p Price) Yen() int { return int(p) }

// --- Availability ---

type Availability bool

const (
	Available   Availability = true
	SoldOut     Availability = false
)
