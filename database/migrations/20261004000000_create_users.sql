-- +goose Up
CREATE TABLE public.users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(200) NOT NULL CHECK (char_length(name) > 0)
);

-- +goose Down
DROP TABLE IF EXISTS public.users;
