-- +goose Up
INSERT INTO app.users(email, password_hash, role)
VALUES ('user@example.com', crypt('user123', gen_salt('bf')), 'user')
    ON CONFLICT (email) DO NOTHING;

-- +goose Down
DELETE FROM app.users WHERE email='user@example.com';
