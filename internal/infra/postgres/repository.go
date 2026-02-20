package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophkeeper/internal/dto/model"
)

func (p *Postgres) SecretExists(ctx context.Context, user, secretName, typeSecret string) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1 
            FROM secret 
            WHERE name = $1 and username = $2 and type = $3
        )
    `
	var exists bool
	err := p.db.QueryRowContext(ctx, query, secretName, user, typeSecret).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error check login %s exists: %w", secretName, err)
	}

	return exists, nil
}

func (r *Postgres) СreateSecret(ctx context.Context, user string, secret *model.Secret) error {
	query := `INSERT INTO secret 
              (username, name, description, type, data) 
              VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		user,
		secret.Name,
		secret.Description,
		secret.SecretType,
		secret.Data,
	)

	if err != nil {
		return fmt.Errorf("error creating login_pass_secrets: %w", err)
	}

	return nil
}

func (r *Postgres) UpdateSecret(ctx context.Context, user string, secret *model.Secret) error {
	query := `UPDATE secret 
              SET name = $1, description = $2, type = $3, data = $4 
              WHERE username = $5 AND name = $1`

	_, err := r.db.ExecContext(ctx, query,
		secret.Name,
		secret.Description,
		secret.SecretType,
		secret.Data,
		user,
	)

	if err != nil {
		return fmt.Errorf("error updating secret: %w", err)
	}

	return nil
}

func (r *Postgres) GetSecret(ctx context.Context, user, secretName, secretType string) (model.Secret, error) {
	var name, description, sType string
	var data []byte

	query := `
            SELECT name, description, type, data
            FROM secret
            WHERE username = $1 and type = $2 and name = $3
        `
	err := r.db.QueryRowContext(ctx, query, user, secretType, secretName).Scan(
		&name,
		&description,
		&sType,
		&data,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Secret{}, fmt.Errorf("secret not found: %s", secretName)
	}
	if err != nil {
		return model.Secret{}, fmt.Errorf("error getting secret: %w", err)
	}
	return model.Secret{
		Name:        name,
		Description: description,
		SecretType:  sType,
		Data:        data,
	}, nil
}

func (r *Postgres) GetSecretList(ctx context.Context, user string) ([]model.Secret, error) {

	query := `
        SELECT name, description, type
		FROM secret
		WHERE username = $1
    `

	rows, err := r.db.QueryContext(ctx, query, user)
	if err != nil {
		return nil, err
	}

	var secrets []model.Secret
	for rows.Next() {
		secret := model.Secret{}
		err = rows.Scan(
			&secret.Name,
			&secret.Description,
			&secret.SecretType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metrics: %w", err)
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

func (r *Postgres) GetSecretListInType(ctx context.Context, user string, typeSecret string) ([]model.Secret, error) {

	query := `
        SELECT name, description, type
		FROM secret
		WHERE username = $1 and type = $2
    `

	rows, err := r.db.QueryContext(ctx, query, user, typeSecret)
	if err != nil {
		return nil, err
	}

	var secrets []model.Secret
	for rows.Next() {
		secret := model.Secret{}
		err = rows.Scan(
			&secret.Name,
			&secret.Description,
			&secret.SecretType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metrics: %w", err)
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

func (r *Postgres) DeleteSecret(ctx context.Context, user, secretName, secretType string) error {

	query := `
        DELETE FROM secret 
		WHERE name = $1 AND username = $2 AND type = $3
    `
	_, err := r.db.ExecContext(ctx, query, secretName, user, secretType)
	if err != nil {
		return err
	}
	return nil
}
