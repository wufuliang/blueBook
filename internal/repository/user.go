package repository

import (
	"blueBook/internal/domain"
	"blueBook/internal/repository/dao"
	"context"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}
func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return repo.toDomain(u), nil
}

func (repo *UserRepository) FindById(ctx context.Context, Id int64) (domain.User, error) {
	u, err := repo.dao.FindById(ctx, Id)
	if err != nil {
		return domain.User{}, err
	}

	return repo.toDomain(u), nil
}
func (repo *UserRepository) UpdateById(ctx context.Context, u domain.User) error {
	return repo.dao.Update(ctx, dao.User{
		Id:       u.Id,
		NickName: u.NickName,
		Birthday: u.Birthday,
		Profile:  u.Profile,
	})
}

func (repo *UserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		NickName: u.NickName,
		Birthday: u.Birthday,
		Profile:  u.Profile,
	}
}
