package service

import (
	"blueBook/internal/domain"
	"blueBook/internal/repository"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
	ErrUserNotFound          = errors.New("用户不存在")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	user, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	return user, nil
}

func (svc *UserService) Edit(ctx context.Context, u domain.User) error {
	// 先查询用户是否存在
	user, err := svc.repo.FindById(ctx, u.Id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}
	user.NickName = u.NickName
	user.Birthday = u.Birthday
	user.Profile = u.Profile
	err = svc.repo.UpdateById(ctx, user)
	return nil
}
