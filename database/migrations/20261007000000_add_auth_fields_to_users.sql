-- +goose Up
ALTER TABLE public.users
    ADD COLUMN login VARCHAR(200),
    ADD COLUMN password_hash TEXT,
    ADD CONSTRAINT users_login_non_empty
        CHECK (login IS NULL OR char_length(btrim(login)) > 0);

CREATE UNIQUE INDEX users_login_unique_idx
    ON public.users (lower(login))
    WHERE login IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS public.users_login_unique_idx;

ALTER TABLE public.users
    DROP CONSTRAINT IF EXISTS users_login_non_empty,
    DROP COLUMN IF EXISTS password_hash,
    DROP COLUMN IF EXISTS login;
