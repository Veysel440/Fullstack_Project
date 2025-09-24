-- 0003_constraints.sql
-- +goose Up
ALTER TABLE app.items
    ADD CONSTRAINT items_price_nonneg CHECK (price >= 0);
-- +goose Down
ALTER TABLE app.items
DROP CONSTRAINT IF EXISTS items_price_nonneg;
