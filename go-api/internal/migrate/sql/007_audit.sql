-- +goose Up
CREATE TABLE IF NOT EXISTS app.item_audit (
                                              id          bigserial PRIMARY KEY,
                                              evt_type    text        NOT NULL,
                                              payload     jsonb       NOT NULL,
                                              created_at  timestamptz NOT NULL DEFAULT now()
    );
CREATE INDEX IF NOT EXISTS idx_item_audit_created_at ON app.item_audit(created_at);

-- +goose Down
DROP TABLE IF EXISTS app.item_audit;
