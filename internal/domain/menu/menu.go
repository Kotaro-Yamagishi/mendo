package menu

import "mendo/internal/domain"

// Menu はメニュー集約のルート。
type Menu struct {
	id           MenuID
	name         MenuName
	price        Price
	availability Availability
}

func NewMenu(id MenuID, name MenuName, price Price) *Menu {
	return &Menu{
		id:           id,
		name:         name,
		price:        price,
		availability: Available,
	}
}

// ReconstructMenu は DB から復元する時に使う。
// NewMenu と違い、ドメインイベントを発行しない。
func ReconstructMenu(id MenuID, name MenuName, price Price, available bool) *Menu {
	availability := SoldOut
	if available {
		availability = Available
	}
	return &Menu{id: id, name: name, price: price, availability: availability}
}

// SoldOut は品切れにする。
func (m *Menu) SoldOut() {
	m.availability = SoldOut
}

func (m *Menu) ID() MenuID       { return m.id }
func (m *Menu) Name() MenuName   { return m.name }
func (m *Menu) Price() Price      { return m.price }
func (m *Menu) IsAvailable() bool { return bool(m.availability) }

func (m *Menu) DomainEvents() []domain.Event {
	return nil
}
