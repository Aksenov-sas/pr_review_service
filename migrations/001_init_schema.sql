-- +goose Up
-- Создание таблиц для сервиса назначения ревьюверов

-- Таблица команд
CREATE TABLE IF NOT EXISTS teams (
    name VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) REFERENCES teams(name) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для поиска пользователей по команде
CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);

-- Таблица pull requests
CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(500) NOT NULL,
    author_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    assigned_reviewers TEXT[], -- массив ID ревьюверов
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP WITH TIME ZONE
);

-- Индекс для поиска PR по автору
CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id ON pull_requests(author_id);

-- Индекс для поиска PR по статусу
CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pull_requests(status);

-- Индекс для поиска PR по назначенным ревьюверам (GIN индекс для массивов)
CREATE INDEX IF NOT EXISTS idx_pull_requests_assigned_reviewers ON pull_requests USING GIN (assigned_reviewers);

-- Триггер для автоматического обновления поля updated_at
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';
-- +goose StatementEnd

-- Применение триггера к таблице пользователей
-- +goose StatementBegin
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO teams (name) VALUES ('default_team') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- Удаление таблиц для сервиса

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;