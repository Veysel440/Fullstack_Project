CREATE TABLE IF NOT EXISTS app.item_audit_parking(
                                                     id bigserial PRIMARY KEY,
                                                     created_at timestamptz NOT NULL DEFAULT now(),
    attempts int NOT NULL DEFAULT 0,
    payload jsonb NOT NULL
    );