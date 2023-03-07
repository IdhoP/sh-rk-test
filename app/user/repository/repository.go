package repository

import (
	"gorm.io/gorm"

	"sh-rk-test/app/user/model"
)

type UserRepository struct {
	Conn gorm.DB
}

func NewUserRepository(conn gorm.DB) *UserRepository {
	return &UserRepository{}
}

func (u *UserRepository) Detail(id int64) (res model.User, err error) {
	// ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	// defer cancel()

	// res, err = u.userRepo.GetByID(ctx, id)
	var result model.User
	flag := u.Conn.First(&result, 1).Error
	if flag != nil {
		err = flag
		return
	}
	res = result
	return
}
