-- +goose Up
CREATE TABLE IF NOT EXISTS public.wallet_transactions (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    amount     BIGINT NOT NULL,
    kind       TEXT NOT NULL CHECK (kind IN ('top_up', 'withdraw')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS wallet_transactions_user_id_idx
    ON public.wallet_transactions (user_id);

-- +goose Down
DROP TABLE IF EXISTS public.wallet_transactions;
