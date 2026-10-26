package repository

import (
	"context"
	"errors"
	"go_marketplace_v1/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrStockOverflow   = errors.New("stock overflow")
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) Create(ctx context.Context, p model.Product) (model.Product, error) {
	const query = `
		INSERT INTO public.products (name, price, stock)
		VALUES ($1, $2, $3)
		RETURNING id, name, price, stock
	`
	err := r.pool.QueryRow(ctx, query, p.Name, p.Price, p.Stock).
		Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	return p, err
}

func (r *ProductRepository) GetAll(ctx context.Context, q string, limit, offset int) ([]model.Product, error) {
	const query = `
		SELECT id, name, price, stock
		FROM public.products
		WHERE name ILIKE '%' || $1 || '%'
		ORDER BY id ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (model.Product, error) {
	const query = `SELECT id, name, price, stock FROM public.products WHERE id = $1`
	var p model.Product
	err := r.pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, ErrProductNotFound
		}
		return model.Product{}, err
	}
	return p, nil
}

func (r *ProductRepository) Update(ctx context.Context, id int64, name *string, price *int64) (model.Product, error) {
	// Сначала получаем текущий товар, чтобы убедиться, что он существует
	p, err := r.GetByID(ctx, id)
	if err != nil {
		return model.Product{}, err
	}

	if name != nil {
		p.Name = *name
	}
	if price != nil {
		p.Price = *price
	}

	const query = `
		UPDATE public.products
		SET name = $1, price = $2, updated_at = now()
		WHERE id = $3
		RETURNING id, name, price, stock
	`
	err = r.pool.QueryRow(ctx, query, p.Name, p.Price, id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	return p, err
}

func (r *ProductRepository) SetStock(ctx context.Context, id int64, stock int64) (model.Product, error) {
	const query = `
		UPDATE public.products
		SET stock = $1, updated_at = now()
		WHERE id = $2
		RETURNING id, name, price, stock
	`
	var p model.Product
	err := r.pool.QueryRow(ctx, query, stock, id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, ErrProductNotFound
		}
		return model.Product{}, err
	}
	return p, nil
}

func (r *ProductRepository) AddStock(ctx context.Context, id int64, quantity int64) (model.Product, error) {
	// Получаем текущий остаток для проверки переполнения (макс. значение INTEGER в PostgreSQL = 2147483647)
	p, err := r.GetByID(ctx, id)
	if err != nil {
		return model.Product{}, err
	}

	if p.Stock+quantity > 2147483647 {
		return model.Product{}, ErrStockOverflow
	}

	const query = `
		UPDATE public.products
		SET stock = stock + $1, updated_at = now()
		WHERE id = $2
		RETURNING id, name, price, stock
	`
	err = r.pool.QueryRow(ctx, query, quantity, id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	return p, err
}
