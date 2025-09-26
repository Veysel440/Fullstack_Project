-- Schema
CREATE SCHEMA IF NOT EXISTS app;


CREATE EXTENSION IF NOT EXISTS citext;

-- Users
CREATE TABLE IF NOT EXISTS app.users (
                                         id            BIGSERIAL PRIMARY KEY,
                                         email         CITEXT UNIQUE NOT NULL,
                                         role          TEXT NOT NULL CHECK (role IN ('user','admin')),
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- Items
CREATE TABLE IF NOT EXISTS app.items (
                                         id          BIGSERIAL PRIMARY KEY,
                                         name        TEXT NOT NULL,
                                         price       NUMERIC(12,2) NOT NULL CHECK (price >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- Indexler (listeleme/sıralama için)
CREATE INDEX IF NOT EXISTS idx_items_created_at ON app.items(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_items_price      ON app.items(price);


--   admin@example.com / password
--   user@example.com  / password
INSERT INTO app.users (email, role, password_hash)
VALUES
    ('admin@example.com','admin','$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcflov8hb7RERJvG3ZIVS4iWf.u'),
    ('user@example.com','user', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcflov8hb7RERJvG3ZIVS4iWf.u')
    ON CONFLICT (email) DO NOTHING;

-- 30 adet sahte ürün
INSERT INTO app.items (name, price, created_at)
SELECT
    'Item ' || g,
    ROUND((random()*900 + 100)::numeric, 2),
    now() - (g || ' days')::interval
FROM generate_series(1, 30) AS g;
