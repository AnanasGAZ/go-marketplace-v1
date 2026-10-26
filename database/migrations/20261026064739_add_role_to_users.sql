-- +goose Up
ALTER TABLE public.users 
ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user' 
CHECK (role IN ('user', 'admin'));

-- +goose Down
ALTER TABLE public.users DROP COLUMN role;
