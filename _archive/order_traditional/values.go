package order

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// --- OrderID ---

type OrderID string

func NewOrderID() OrderID        { return OrderID(uuid.New().String()) }
func (id OrderID) String() string { return string(id) }

// --- SeatNumber ---

type SeatNumber int

func NewSeatNumber(n int) (SeatNumber, error) {
	if n < 1 || n > 50 {
		return 0, fmt.Errorf("座席番号は1-50の範囲です: %d", n)
	}
	return SeatNumber(n), nil
}

// --- Topping ---

type Topping string

const (
	ToppingAjitama   Topping = "味玉"
	ToppingChashu    Topping = "チャーシュー"
	ToppingNori      Topping = "海苔"
	ToppingMenma     Topping = "メンマ"
	ToppingCorn      Topping = "コーン"
	ToppingButter    Topping = "バター"
)

func ParseTopping(s string) (Topping, error) {
	switch Topping(s) {
	case ToppingAjitama, ToppingChashu, ToppingNori, ToppingMenma, ToppingCorn, ToppingButter:
		return Topping(s), nil
	default:
		return "", fmt.Errorf("無効なトッピング: %s", s)
	}
}

// --- CookingNote ---

type CookingNote struct {
	noodleHardness string // かため、ふつう、やわらかめ
	extra          string // ネギ多め、等
}

func NewCookingNote(hardness, extra string) (CookingNote, error) {
	valid := map[string]bool{"かため": true, "ふつう": true, "やわらかめ": true}
	if !valid[hardness] {
		return CookingNote{}, errors.New("麺の硬さは かため/ふつう/やわらかめ のいずれか")
	}
	return CookingNote{noodleHardness: hardness, extra: extra}, nil
}

func (c CookingNote) NoodleHardness() string { return c.noodleHardness }
func (c CookingNote) Extra() string          { return c.extra }

// --- Status ---

type Status int

const (
	StatusPending Status = iota
	StatusConfirmed
	StatusCancelled
)
