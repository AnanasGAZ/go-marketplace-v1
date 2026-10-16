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

var (
	ErrInvalidAmount   = errors.New("сумма должна быть больше нуля")
	ErrBalanceOverflow = errors.New("переполнение баланса")
)

const maxBalance = int64(1) << 62 // ~4.6e18, безопасный потолок для int64

func (s *WalletService) TopUp(ctx context.Context, userID, amount int64) (model.Wallet, int64, error) {
	if amount <= 0 {
		return model.Wallet{}, 0, ErrInvalidAmount
	}

	// Проверяем, что пользователь существует
	if _, err := s.userRepository.GetByID(ctx, userID); err != nil {
		// тут та же проверка, что в GetByUserID
		return model.Wallet{}, 0, err
	}

	// Проверяем переполнение ДО обращения к БД
	wallet, err := s.walletRepository.GetByUserID(ctx, userID)
	if err != nil {
		return model.Wallet{}, 0, err
	}
	if wallet.Balance > maxBalance-amount {
		return model.Wallet{}, 0, ErrBalanceOverflow
	}

	updated, transactionID, err := s.walletRepository.TopUp(ctx, userID, amount)
	if err != nil {
		return model.Wallet{}, 0, err
	}

	return updated, transactionID, nil
}
