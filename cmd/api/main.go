package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go_marketplace_v1/internal/config"
	"go_marketplace_v1/internal/handler"
	"go_marketplace_v1/internal/repository"
	"go_marketplace_v1/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("environment: не удалось загрузить .env: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	//context.Context — это объект, который Go-функции передают друг другу,
	// чтобы сообщить: операцию пора остановить — например,
	// потому что истёк срок ожидания или отменился запрос.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //лимит 5 сек
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL) //менеджерсоединений
	if err != nil {
		return fmt.Errorf("database: не удалось создать пул подключений: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil { // проверим что бд доступна
		return fmt.Errorf("database: не удалось подключиться к PostgreSQL: %w", err)
	}

	userRepo := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	authService, err := service.NewAuthService(userRepo, []byte(cfg.JWTSecret), cfg.JWTTokenTTL)
	if err != nil {
		return fmt.Errorf("auth: не удалось создать сервис авторизации: %w", err)
	}
	authHandler := handler.NewAuthHandler(authService)

	mux := http.NewServeMux() //маршрутизатор
	mux.HandleFunc("POST /api/v1/users", userHandler.Create)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server: %w", err)
	}

	return nil
}
