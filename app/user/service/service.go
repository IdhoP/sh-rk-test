package service

import (
	"sh-rk-test/app/user/model"
	"sh-rk-test/app/user/repository"
)

type UserService struct {
	UserRepo *repository.UserRepository
}

func NewUserService(ur *repository.UserRepository) *UserService {
	return &UserService{
		UserRepo: ur,
	}
}

func (u *UserService) Detail(id int64) (res model.User, err error) {
	// ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	// defer cancel()

	// res, err = u.userRepo.GetByID(ctx, id)
	res, err = u.UserRepo.Detail(id)
	if err != nil {
		return
	}
	return
}
