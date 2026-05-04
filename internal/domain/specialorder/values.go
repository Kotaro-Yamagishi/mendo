package specialorder

import "github.com/google/uuid"

type SpecialOrderID string

func NewSpecialOrderID() SpecialOrderID { return SpecialOrderID(uuid.New().String()) }
func (id SpecialOrderID) String() string { return string(id) }

type SpecialOrderStatus int

const (
	StatusRequested  SpecialOrderStatus = iota // 申請中
	StatusPending                               // 承認待ち
	StatusApproved                              // 承認済み
	StatusRejected                              // 却下
	StatusCooking                               // 調理中
	StatusCompleted                             // 完了
)

func (s SpecialOrderStatus) String() string {
	switch s {
	case StatusRequested:
		return "requested"
	case StatusPending:
		return "pending"
	case StatusApproved:
		return "approved"
	case StatusRejected:
		return "rejected"
	case StatusCooking:
		return "cooking"
	case StatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}
