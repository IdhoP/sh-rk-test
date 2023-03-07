package model

import (
	"time"

	"gorm.io/gorm"
)

type OrderItem struct {
	gorm.Model
	Name      string
	Price     int64
	ExpiredAt time.Time
}

type OrderItemPayload struct {
	Name      string    `form:"name" json:"name" validate:"required"`
	Price     int64     `form:"price" json:"price" validate:"required,numeric"`
	ExpiredAt time.Time `form:"expired_at" json:"expired_at" validate:"required"`
}
