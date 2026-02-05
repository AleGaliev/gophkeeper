-- Таблица для хранения пользователей
CREATE TABLE users (
                       login VARCHAR(255) NOT NULL UNIQUE,
                       password_hash VARCHAR(255) NOT NULL,
                       public_key TEXT,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица для хранения логинов и паролей
CREATE TABLE login_pass_secrets (
                                    username VARCHAR(255) NOT NULL,
                                    name VARCHAR(255) NOT NULL,
                                    description TEXT,
                                    login TEXT NOT NULL,
                                    password TEXT NOT NULL,
                                    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица для хранения текстовых секретов
CREATE TABLE text_secrets (
                              username VARCHAR(255) NOT NULL,
                              name VARCHAR(255) NOT NULL,
                              description TEXT,
                              text TEXT NOT NULL,
                              created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                              updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица для хранения бинарных данных
CREATE TABLE binary_secrets (
                                username VARCHAR(255) NOT NULL,
                                name VARCHAR(255) NOT NULL,
                                description TEXT,
                                data BYTEA NOT NULL,
                                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица для хранения данных банковских карт
CREATE TABLE bank_card_secrets (
                                   username VARCHAR(255) NOT NULL,
                                   name VARCHAR(255) NOT NULL,
                                   description TEXT,
                                   card_number TEXT NOT NULL,
                                   expiry_month TEXT NOT NULL,
                                   expiry_year TEXT NOT NULL,
                                   cvv TEXT NOT NULL,
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

-- Триггер для таблицы login_pass_secrets
CREATE TRIGGER update_login_pass_secrets_updated_at
    BEFORE UPDATE ON login_pass_secrets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Триггер для таблицы text_secrets
CREATE TRIGGER update_text_secrets_updated_at
    BEFORE UPDATE ON text_secrets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Триггер для таблицы binary_secrets
CREATE TRIGGER update_binary_secrets_updated_at
    BEFORE UPDATE ON binary_secrets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Триггер для таблицы bank_card_secrets
CREATE TRIGGER update_bank_card_secrets_updated_at
    BEFORE UPDATE ON bank_card_secrets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();