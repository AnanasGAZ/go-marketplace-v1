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
func (r *WalletRepository) TopUp(ctx context.Context, userID, amount int64) (model.Wallet, int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Wallet{}, 0, err
	}
	defer tx.Rollback(ctx) // если не закоммитим — откатится

	const updateBalance = `
		UPDATE public.wallets
		SET balance = balance + $1,
		    updated_at = now()
		WHERE user_id = $2
		RETURNING user_id, balance
	`

	var wallet model.Wallet
	if err := tx.QueryRow(ctx, updateBalance, amount, userID).Scan(&wallet.UserID, &wallet.Balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Wallet{}, 0, ErrWalletNotFound
		}
		return model.Wallet{}, 0, err
	}

	const insertTx = `
		INSERT INTO public.wallet_transactions (user_id, amount, kind)
		VALUES ($1, $2, 'top_up')
		RETURNING id
	`

	var transactionID int64
	if err := tx.QueryRow(ctx, insertTx, userID, amount).Scan(&transactionID); err != nil {
		return model.Wallet{}, 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Wallet{}, 0, err
	}

	return wallet, transactionID, nil
}
