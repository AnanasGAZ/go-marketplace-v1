package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"go_marketplace_v1/internal/model"
	"go_marketplace_v1/internal/repository"
)

var (
	ErrInvalidCredentials  = errors.New("неверный логин или пароль")
	ErrInvalidRegistration = errors.New("логин и пароль обязательны")
	ErrLoginAlreadyExists  = errors.New("логин уже занят")
	ErrInvalidJWTSecret    = errors.New("секрет JWT должен содержать не менее 32 байт")
	ErrInvalidTokenTTL     = errors.New("срок действия JWT должен быть положительным")
)

type AuthService struct {
	userRepository *repository.UserRepository
	jwtSecret      []byte
	tokenTTL       time.Duration
}

func NewAuthService(userRepository *repository.UserRepository, jwtSecret []byte, tokenTTL time.Duration) (*AuthService, error) {
	if userRepository == nil {
		return nil, errors.New("репозиторий пользователей не задан")
	}
	if len(jwtSecret) < 32 {
		return nil, ErrInvalidJWTSecret
	}
	if tokenTTL <= 0 {
		return nil, ErrInvalidTokenTTL
	}

	return &AuthService{
		userRepository: userRepository,
		jwtSecret:      append([]byte(nil), jwtSecret...),
		tokenTTL:       tokenTTL,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, name, login, password string) (model.User, error) {
	name = strings.TrimSpace(name)
	if length := utf8.RuneCountInString(name); length < 1 || length > 200 {
		return model.User{}, ErrInvalidName
	}

	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return model.User{}, ErrInvalidRegistration
	}

	passwordHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return model.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepository.CreateWithCredentials(ctx, name, login, passwordHash)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) &&
			postgresError.Code == "23505" &&
			postgresError.ConstraintName == "users_login_unique_idx" {
			return model.User{}, ErrLoginAlreadyExists
		}
		return model.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	user, err := s.userRepository.GetByLogin(ctx, login)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrInvalidCredentials
	}
	if err != nil {
		return "", fmt.Errorf("find user by login: %w", err)
	}
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return "", ErrInvalidCredentials
	}

	passwordMatches, err := argon2id.ComparePasswordAndHash(password, *user.PasswordHash)
	if err != nil {
		return "", fmt.Errorf("compare password hash: %w", err)
	}
	if !passwordMatches {
		return "", ErrInvalidCredentials
	}

	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(user.ID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	return signedToken, nil
}
