package order

import (
	"fmt"

	"github.com/google/uuid"

	"mendo/internal/apperrors"
)

// --- OrderID ---

type OrderID string

func NewOrderID() OrderID         { return OrderID(uuid.New().String()) }
func (id OrderID) String() string { return string(id) }

// --- SeatNumber ---

// SeatNumber は座席番号。1-100 の範囲。
type SeatNumber int

func NewSeatNumber(n int) (SeatNumber, error) {
	if n < 1 || n > 100 {
		return 0, apperrors.Domain(ErrCodeInvalidSeatNumber, fmt.Sprintf("座席番号は1〜100の範囲で指定してください: %d", n))
	}
	return SeatNumber(n), nil
}

// --- Hardness ---

// Hardness は麺の硬さ。
type Hardness string

const (
	HardnessKatame     Hardness = "かため"
	HardnessFutsuu     Hardness = "ふつう"
	HardnessYawarakame Hardness = "やわらかめ"
)

func NewHardness(s string) (Hardness, error) {
	switch Hardness(s) {
	case HardnessKatame, HardnessFutsuu, HardnessYawarakame:
		return Hardness(s), nil
	default:
		return "", apperrors.Domain(ErrCodeInvalidHardness, fmt.Sprintf("麺の硬さは「かため」「ふつう」「やわらかめ」のいずれかです: %s", s))
	}
}

// --- Status ---

const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCanceled  = "canceled"
)
