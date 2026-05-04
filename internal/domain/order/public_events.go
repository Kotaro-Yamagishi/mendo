package order

import "mendo/internal/domain/contract"

// 内部イベント → 公開イベントへの変換関数。
// 型定義は contract/ に移動済み。ここには変換関数だけ残す。
// 内部イベントの構造を知っているのはこの BC だけ。変換もここで行う。

// ToPublicConfirmed は内部イベントを公開イベントに変換する。
func ToPublicConfirmed(internal OrderConfirmed) contract.OrderConfirmedPublic {
	items := make([]contract.OrderConfirmedPublicItem, len(internal.Items))
	for i, item := range internal.Items {
		items[i] = contract.OrderConfirmedPublicItem{
			MenuName: item.MenuID, // 本番では MenuID → メニュー名に変換するが学習用なので ID のまま
			Toppings: item.Toppings,
			Hardness: item.Hardness,
		}
	}
	return contract.OrderConfirmedPublic{
		OrderID: internal.GetAggregateID(),
		SeatNo:  internal.SeatNo,
		Items:   items,
	}
}

// ToPublicCanceled は内部イベントを公開イベントに変換する。
func ToPublicCanceled(internal OrderCancelled) contract.OrderCanceledPublic {
	return contract.OrderCanceledPublic{
		OrderID: internal.GetAggregateID(),
		Reason:  internal.Reason,
	}
}

// ToPublicCreated は内部イベントを公開イベントに変換する。
func ToPublicCreated(internal OrderCreated) contract.OrderCreatedPublic {
	return contract.OrderCreatedPublic{
		OrderID: internal.GetAggregateID(),
		SeatNo:  internal.SeatNo,
	}
}
