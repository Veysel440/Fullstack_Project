-- +goose Up
INSERT INTO app.users(email, password_hash, role)
VALUES ('admin@example.com', crypt('admin123', gen_salt('bf')), 'admin')
    ON CONFLICT (email) DO NOTHING;

-- +goose Down
DELETE FROM app.users WHERE email='admin@example.com';
