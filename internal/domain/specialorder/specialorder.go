package specialorder

import (
	"errors"

	commondomain "mendo/internal/domain"
)

// SpecialOrder はプロセスマネージャー。
// サーガと違い「今どのステップか」を状態として持ち、次のアクションを判断する。
type SpecialOrder struct {
	id            SpecialOrderID
	orderID       string
	menuName      string
	status        SpecialOrderStatus
	suggestedMenu string
	domainEvents  []commondomain.Event
}

func NewSpecialOrder(id SpecialOrderID, orderID, menuName string) *SpecialOrder {
	so := &SpecialOrder{
		id:       id,
		orderID:  orderID,
		menuName: menuName,
		status:   StatusPending,
	}
	so.domainEvents = append(so.domainEvents, NewSpecialOrderRequested(id, orderID, menuName, ""))
	return so
}

// Approve は店長が承認する。承認後、自動的に調理開始を判断する（プロセスマネージャーの役割）。
func (so *SpecialOrder) Approve() error {
	if so.status != StatusPending {
		return errors.New("承認待ち状態のみ承認できます")
	}
	so.status = StatusApproved
	so.domainEvents = append(so.domainEvents, NewSpecialOrderApproved(so.id, ""))

	// プロセスマネージャーが次のステップを判断: 承認 → 調理開始
	so.status = StatusCooking
	so.domainEvents = append(so.domainEvents, NewCookingDispatched(so.id, so.orderID, ""))
	return nil
}

// Reject は店長が却下する。
func (so *SpecialOrder) Reject(reason, suggestedMenu string) error {
	if so.status != StatusPending {
		return errors.New("承認待ち状態のみ却下できます")
	}
	so.status = StatusRejected
	so.suggestedMenu = suggestedMenu
	so.domainEvents = append(so.domainEvents, NewSpecialOrderRejected(so.id, reason, suggestedMenu, ""))
	return nil
}

// ResubmitWithMenu は却下後に別メニューで再申請する。再度承認待ちに戻る。
func (so *SpecialOrder) ResubmitWithMenu(newMenuName string) error {
	if so.status != StatusRejected {
		return errors.New("却下済みのみ再申請できます")
	}
	so.menuName = newMenuName
	so.status = StatusPending
	so.domainEvents = append(so.domainEvents, NewMenuResubmitted(so.id, newMenuName, ""))
	return nil
}

func (so *SpecialOrder) ID() SpecialOrderID                { return so.id }
func (so *SpecialOrder) OrderID() string                   { return so.orderID }
func (so *SpecialOrder) MenuName() string                  { return so.menuName }
func (so *SpecialOrder) Status() SpecialOrderStatus        { return so.status }
func (so *SpecialOrder) SuggestedMenu() string             { return so.suggestedMenu }
func (so *SpecialOrder) DomainEvents() []commondomain.Event { return so.domainEvents }

// ReconstructSpecialOrder は DB から読み込んだ値を使って SpecialOrder を復元する。
// ドメインイベントは発行しない（永続化済みの状態を読み戻すだけ）。
// infrastructure 層のリポジトリ実装専用。
func ReconstructSpecialOrder(id SpecialOrderID, orderID, menuName string, status SpecialOrderStatus, suggestedMenu string) *SpecialOrder {
	return &SpecialOrder{
		id:            id,
		orderID:       orderID,
		menuName:      menuName,
		status:        status,
		suggestedMenu: suggestedMenu,
	}
}
