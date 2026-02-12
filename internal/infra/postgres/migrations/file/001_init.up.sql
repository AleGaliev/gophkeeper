-- Таблица для хранения пользователей
CREATE TABLE users (
                       login VARCHAR(255) NOT NULL UNIQUE,
                       public_key TEXT,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE secret (
                       username VARCHAR(255) NOT NULL,
                       name VARCHAR(255) NOT NULL,
                       description TEXT,
                       type VARCHAR(255) NOT NULL,
                       data BYTEA NOT NULL,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Функция для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для таблицы users
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Триггер для таблицы secret
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON secret
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
