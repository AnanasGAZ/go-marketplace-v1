package repository

import (
	"context"

	"go_marketplace_v1/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// пул  подключений к бд
type UserRepository struct {
	pool *pgxpool.Pool //указатель на пул
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool} // берёт адрес структуры, чтобы вернуть указатель.
}

func (r *UserRepository) Create(ctx context.Context, name string) (model.User, error) {
	const query = `
		INSERT INTO public.users (name)
		VALUES ($1)
		RETURNING id, name
	`

	var user model.User
	if err := r.pool.QueryRow(ctx, query, name).Scan(&user.ID, &user.Name); err != nil {
		return model.User{}, err
	}

	return user, nil
}
