DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_login_pass_secrets_updated_at ON login_pass_secrets;
DROP TRIGGER IF EXISTS update_text_secrets_updated_at ON text_secrets;
DROP TRIGGER IF EXISTS update_binary_secrets_updated_at ON binary_secrets;
DROP TRIGGER IF EXISTS update_bank_card_secrets_updated_at ON bank_card_secrets;

-- Удаление функции
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаление таблиц в обратном порядке (сначала дочерние, потом родительские)
DROP TABLE IF EXISTS bank_card_secrets;
DROP TABLE IF EXISTS binary_secrets;
DROP TABLE IF EXISTS text_secrets;
DROP TABLE IF EXISTS login_pass_secrets;
DROP TABLE IF EXISTS users;