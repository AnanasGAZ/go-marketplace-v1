-- +goose Up
CREATE TABLE IF NOT EXISTS public.wallets (
    user_id    BIGINT PRIMARY KEY REFERENCES public.users(id) ON DELETE CASCADE,
    balance    BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose Down
DROP TABLE IF EXISTS public.wallets;