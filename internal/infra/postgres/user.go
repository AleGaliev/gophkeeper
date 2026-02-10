package postgres

import (
	"context"
	"fmt"
	"gophkeeper/internal/dto/model"
)

func (p *Postgres) CreateUser(ctx context.Context, user model.User) error {
	query := `
	INSERT INTO users (login, password_hash)
	VALUES ($1, $2)
	`
	_, err := p.db.ExecContext(ctx, query, user.Login, user.Hash)
	if err != nil {
		return fmt.Errorf("error create user %s: %w", user.Login, err)
	}
	return nil
}

func (p *Postgres) UserExists(ctx context.Context, login string) (bool, error) {
	query := `
	SELECT EXISTS(
		SELECT 1 
		FROM users 
		WHERE login = $1
	)
	`

	var exists bool
	err := p.db.QueryRowContext(ctx, query, login).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error check user %s exists: %w", login, err)
	}

	return exists, nil
}

func (p *Postgres) GetUserHash(ctx context.Context, login string) (string, error) {
	var passwordHash string
	query := `SELECT password_hash FROM users WHERE login = $1`
	if err := p.db.QueryRowContext(ctx, query, login).Scan(&passwordHash); err != nil {
		return passwordHash, fmt.Errorf("failed to check user existence: %v", err)
	}
	return passwordHash, nil
}
