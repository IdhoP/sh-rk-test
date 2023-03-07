package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FullName   string
	FirstOrder string
}

type UserPayload struct {
	FullName   string `form:"full_name" json:"full_name" validate:"required"`
	FirstOrder string `form:"first_order" json:"first_order" validate:"required"`
}
