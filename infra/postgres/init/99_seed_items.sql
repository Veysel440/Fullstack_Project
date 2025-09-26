-- Şema ve tablo
CREATE SCHEMA IF NOT EXISTS app;

CREATE TABLE IF NOT EXISTS app.items (
                                         id         BIGSERIAL PRIMARY KEY,
                                         name       TEXT NOT NULL,
                                         price      NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- Yeni alanlar (geriye uyumlu)
ALTER TABLE app.items
    ADD COLUMN IF NOT EXISTS category TEXT    NOT NULL DEFAULT 'general',
    ADD COLUMN IF NOT EXISTS stock    INTEGER NOT NULL DEFAULT 0;

-- Faydalı indexler
CREATE INDEX IF NOT EXISTS idx_items_category ON app.items(category);
CREATE INDEX IF NOT EXISTS idx_items_price    ON app.items(price);



-- 500 sahte ürün
WITH gen AS (
    SELECT
        'Item ' || gs                                       AS name,
        ROUND((random()*990 + 10)::numeric, 2)              AS price,
        (ARRAY['electronics','books','toys','grocery','clothes'])
   [1 + floor(random()*5)]::text                     AS category,
    floor(random()*500)::int                            AS stock,
    now() - (floor(random()*90) || ' days')::interval   AS created_at
FROM generate_series(1, 500) AS gs
    )
INSERT INTO app.items (name, price, category, stock, created_at)
SELECT name, price, category, stock, created_at FROM gen;
