package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"go_marketplace_v1/internal/model"
	"go_marketplace_v1/internal/repository"
)

var ErrInvalidName = errors.New("имя должно содержать от 1 до 200 символов")
var ErrUserNotFound = errors.New("user not found")

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(userRepository *repository.UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) Create(ctx context.Context, name string) (model.User, error) {
	name = strings.TrimSpace(name)
	if length := utf8.RuneCountInString(name); length < 1 || length > 200 {
		return model.User{}, ErrInvalidName
	}

	return s.userRepository.Create(ctx, name)
}

// func (s *UserService) GetByID(ctx context.Context, id int64) (model.User, error) {
// 	return s.userRepository.GetByID(ctx, id)
// }

func (s *UserService) GetByID(ctx context.Context, id int64) (model.User, error) {
	user, err := s.userRepository.GetByID(ctx, id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}
