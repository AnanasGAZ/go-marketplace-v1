package service

import (
	"context"
	"errors"

	"go_marketplace_v1/internal/model"
	"go_marketplace_v1/internal/repository"
)

var (
	ErrWalletNotFound   = errors.New("кошелёк не найден")
	ErrUserDoesNotExist = errors.New("пользователь не существует")
)

type WalletService struct {
	walletRepository *repository.WalletRepository
	userRepository   *repository.UserRepository
}

func NewWalletService(
	walletRepository *repository.WalletRepository,
	userRepository *repository.UserRepository,
) *WalletService {
	return &WalletService{
		walletRepository: walletRepository,
		userRepository:   userRepository,
	}
}

func (s *WalletService) GetByUserID(ctx context.Context, userID int64) (model.Wallet, error) {
	// Сначала убеждаемся, что пользователь существует — иначе 404.
	if _, err := s.userRepository.GetByID(ctx, userID); err != nil {
		if errors.Is(err, ErrUserDoesNotExist) {
			return model.Wallet{}, ErrUserDoesNotExist
		}
		return model.Wallet{}, err
	}

	wallet, err := s.walletRepository.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrWalletNotFound) {
			// кошелька нет — создаём
			created, createErr := s.walletRepository.Create(ctx, userID)
			if createErr != nil {
				return model.Wallet{}, createErr
			}
			return created, nil
		}
		return model.Wallet{}, err
	}

	return wallet, nil
}
