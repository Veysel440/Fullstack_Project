-- Varsayılan postgres DB içinde çalışır
CREATE USER app WITH PASSWORD 'app';
CREATE DATABASE appdb OWNER app;
GRANT ALL PRIVILEGES ON DATABASE appdb TO app;
