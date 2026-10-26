package repository

import (
	"context"
	"errors"

	"go_marketplace_v1/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

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
		RETURNING id, name, role
	`

	var user model.User
	if err := r.pool.QueryRow(ctx, query, name).Scan(&user.ID, &user.Name); err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (r *UserRepository) CreateWithCredentials(ctx context.Context, name, login, passwordHash string) (model.User, error) {
	const query = `
		INSERT INTO public.users (name, login, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, login, password_hash, role
	`

	var user model.User
	if err := r.pool.QueryRow(ctx, query, name, login, passwordHash).Scan(
		&user.ID,
		&user.Name,
		&user.Login,
		&user.PasswordHash,
		&user.Role,
	); err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (model.User, error) {
	const query = `
		SELECT id, login, password_hash, role
		FROM public.users
		WHERE lower(login) = lower($1)
	`

	var user model.User
	if err := r.pool.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.Role); err != nil {
		return model.User{}, err
	}

	return user, nil
}

// func (r *UserRepository) GetByID(ctx context.Context, id int64) (model.User, error) {
// 	const query = `
// 		SELECT id, name, login, password_hash
// 		FROM public.users
// 		WHERE id = $1
// 	`

// 	var user model.User

// 	if err := r.pool.QueryRow(ctx, query, id).Scan(
// 		&user.ID,
// 		&user.Name,
// 		&user.Login,
// 		&user.PasswordHash,
// 	); err != nil {
// 		log.Printf("rep: ошибка user id не получил: %v", err)
// 		return model.User{}, err

// 	}

//		return user, nil
//	}
func (r *UserRepository) GetByID(ctx context.Context, id int64) (model.User, error) {
	const query = `
  SELECT id, name, login, password_hash, role
  FROM public.users
  WHERE id = $1
 `

	var user model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Login,
		&user.PasswordHash,
		&user.Role,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}
