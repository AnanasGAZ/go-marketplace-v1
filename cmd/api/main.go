package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"go_marketplace_v1/internal/config"
	"go_marketplace_v1/internal/handler"
	"go_marketplace_v1/internal/repository"
	"go_marketplace_v1/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()

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

	mux := http.NewServeMux() //маршрутизатор
	mux.HandleFunc("POST /api/v1/users", userHandler.Create)

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
