package model

import (
	_orderItem "sh-rk-test/app/order_item/model"
	_user "sh-rk-test/app/user/model"

	"gorm.io/gorm"
)

type OrderHistory struct {
	gorm.Model
	Descriptions string
	UserID       int
	User         _user.User
	OrderItemID  int
	OrderItem    _orderItem.OrderItem
}

type OrderHistoryPayload struct {
	Descriptions string `form:"descriptions" json:"descriptions" validate:"required"`
	UserID       int    `form:"user_id" json:"user_id" validate:"required,numeric"`
	OrderItemID  int    `form:"item_id" json:"item_id" validate:"required,numeric"`
}
