package service

import (
	"context"
	"errors"
	"strings"

	"go_marketplace_v1/internal/model"
	"go_marketplace_v1/internal/repository"
)

var (
	// ErrInvalidName     = errors.New("invalid name: must be 1-200 characters after trimming")
	ErrInvalidPrice    = errors.New("invalid price: must be > 0")
	ErrInvalidStock    = errors.New("invalid stock: must be >= 0")
	ErrInvalidQuantity = errors.New("invalid quantity: must be > 0")
	ErrEmptyUpdate     = errors.New("at least one field (name or price) must be provided")
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func validateProduct(name string, price, stock int64) error {
	name = strings.TrimSpace(name)
	if len(name) < 1 || len(name) > 200 {
		return ErrInvalidName
	}
	if price <= 0 {
		return ErrInvalidPrice
	}
	if stock < 0 {
		return ErrInvalidStock
	}
	return nil
}
func (s *ProductService) Create(ctx context.Context, name string, price, stock int64) (model.Product, error) {
	if err := validateProduct(name, price, stock); err != nil {
		return model.Product{}, err
	}
	p := model.Product{Name: strings.TrimSpace(name), Price: price, Stock: stock}
	return s.repo.Create(ctx, p)
}

func (s *ProductService) GetAll(ctx context.Context, q string, limit, offset int) ([]model.Product, error) {
	return s.repo.GetAll(ctx, q, limit, offset)
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) Update(ctx context.Context, id int64, name *string, price *int64) (model.Product, error) {
	if name == nil && price == nil {
		return model.Product{}, ErrEmptyUpdate
	}
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if len(trimmed) < 1 || len(trimmed) > 200 {
			return model.Product{}, ErrInvalidName
		}
		*name = trimmed
	}
	if price != nil && *price <= 0 {
		return model.Product{}, ErrInvalidPrice
	}
	return s.repo.Update(ctx, id, name, price)
}

func (s *ProductService) SetStock(ctx context.Context, id int64, stock int64) (model.Product, error) {
	if stock < 0 {
		return model.Product{}, ErrInvalidStock
	}
	return s.repo.SetStock(ctx, id, stock) // проверить что не больше max!!!!
}

func (s *ProductService) AddStock(ctx context.Context, id int64, quantity int64) (model.Product, error) {
	if quantity <= 0 {
		return model.Product{}, ErrInvalidQuantity
	}
	return s.repo.AddStock(ctx, id, quantity)
}
