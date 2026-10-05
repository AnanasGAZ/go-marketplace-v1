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
