package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go_marketplace_v1/internal/model"
)

var ErrWalletNotFound = errors.New("wallet not found")

type WalletRepository struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

func (r *WalletRepository) GetByUserID(ctx context.Context, userID int64) (model.Wallet, error) {
	const query = `
		SELECT user_id, balance
		FROM public.wallets
		WHERE user_id = $1
	`

	var wallet model.Wallet
	err := r.pool.QueryRow(ctx, query, userID).Scan(&wallet.UserID, &wallet.Balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Wallet{}, ErrWalletNotFound
		}
		return model.Wallet{}, err
	}

	return wallet, nil
}
func (r *WalletRepository) Create(ctx context.Context, userID int64) (model.Wallet, error) {
	const query = `
		INSERT INTO public.wallets (user_id, balance)
		VALUES ($1, 0)
		RETURNING user_id, balance
	`

	var wallet model.Wallet
	err := r.pool.QueryRow(ctx, query, userID).Scan(&wallet.UserID, &wallet.Balance)
	if err != nil {
		return model.Wallet{}, err
	}

	return wallet, nil
}
