package order

import "mendo/internal/domain/menu"

// OrderItem は注文集約の内部エンティティ。外部から直接操作できない。
type OrderItem struct {
	menuID      menu.MenuID
	toppings    []Topping
	cookingNote CookingNote
}

func NewOrderItem(menuID menu.MenuID, toppings []Topping, note CookingNote) OrderItem {
	return OrderItem{
		menuID:      menuID,
		toppings:    toppings,
		cookingNote: note,
	}
}

func (i OrderItem) MenuID() menu.MenuID    { return i.menuID }
func (i OrderItem) Toppings() []Topping     { return i.toppings }
func (i OrderItem) CookingNote() CookingNote { return i.cookingNote }
