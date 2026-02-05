package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophkeeper/internal/dto/model"
)

//const (
//	tableNameLoginPass = "login_pass_secrets"
//	tableNameBankCard = "bank_cards_secrets"
//	tableNameText = "text_secrets"
//	tableNameBinary = "binary_secrets"
//)

func (r *Postgres) СreateLoginPass(ctx context.Context, user string, secret *model.Secret) error {
	query := `INSERT INTO login_pass_secrets 
              (username, name, description, login, password) 
              VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		user,
		secret.Name,
		secret.Description,
		secret.LoginPass.Login,
		secret.LoginPass.Password,
	)

	if err != nil {
		return fmt.Errorf("error creating login_pass_secrets: %w", err)
	}

	return nil
}

func (r *Postgres) СreateSecretText(ctx context.Context, user string, secret *model.Secret) error {
	query := `INSERT INTO text_secrets 
              (username, name, description, text) 
              VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, user,
		secret.Name,
		secret.Description,
		secret.SecretText.Text,
	)

	if err != nil {
		return fmt.Errorf("error creating text secrets: %w", err)
	}
	return nil
}

func (r *Postgres) CreateBinaryData(ctx context.Context, user string, secret *model.Secret) error {
	query := `INSERT INTO binary_secrets 
              (username, name, description, binary) 
              VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query,
		user,
		secret.Name,
		secret.Description,
		secret.BinaryData.Data,
	)

	if err != nil {
		return fmt.Errorf("error creating secret binary file: %w", err)
	}
	return nil
}

func (r *Postgres) CreateBankCard(ctx context.Context, user string, secret *model.Secret) error {
	query := `INSERT INTO bank_card_secrets 
              (username, name, description, card_number, expiry_month, expiry_year, cvv) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		user,
		secret.Name,
		secret.Description,
		secret.BankCard.CardNumber,
		secret.BankCard.ExpiryMonth,
		secret.BankCard.ExpiryYear,
		secret.BankCard.CVV)
	if err != nil {
		return fmt.Errorf("error creating secret bank card: %w", err)
	}
	return nil
}

func (p *Postgres) SecretExists(user, secretName, typeSecret string) (bool, error) {
	tableName, err := typeSecretInTableName(typeSecret)
	if err != nil {
		return false, fmt.Errorf("unknown secret type: %s", typeSecret)
	}

	query := fmt.Sprintf(`
        SELECT EXISTS(
            SELECT 1 
            FROM %s 
            WHERE name = $1 and username = $2
        )
    `, tableName)

	var exists bool
	err = p.db.QueryRow(query, secretName, user).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error check login %s exists: %w", secretName, err)
	}

	return exists, nil
}

func (r *Postgres) GetSecretList(user string, typeSecret string) (*model.Secrets, error) {

	secrets := model.Secrets{
		SecretType: typeSecret,
		Secrets:    []model.Secret{},
	}

	tableName, err := typeSecretInTableName(typeSecret)
	if err != nil {
		return &secrets, fmt.Errorf("unknown secret type: %s", typeSecret)
	}

	query := fmt.Sprintf(`
        SELECT name, description
		FROM %s
		WHERE username = $1
    `, tableName)

	rows, err := r.db.Query(query, user)
	if err != nil {
		return &secrets, err
	}

	for rows.Next() {
		secret := model.Secret{}
		err = rows.Scan(
			&secret.Name,
			&secret.Description,
		)
		if err != nil {
			return &secrets, fmt.Errorf("failed to scan metrics: %w", err)
		}
		secrets.Secrets = append(secrets.Secrets, secret)
	}

	return &secrets, nil
}

func (r *Postgres) GetBankCard(ctx context.Context, user, secretName string) (*model.Secret, error) {
	var name, description, cardNumber, expiryMonth, expiryYear, CVV string

	query := `
			SELECT name, description, card_number, expiry_month, expiry_year, cvv, metadata, data
			FROM bank_card_secrets
			WHERE username = $1
			`
	err := r.db.QueryRowContext(ctx, query, user).Scan(
		&name,
		&description,
		&cardNumber,
		&expiryMonth,
		&expiryYear,
		&CVV,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return &model.Secret{}, fmt.Errorf("Secret not found: %s", secretName)
	}
	if err != nil {
		return &model.Secret{}, fmt.Errorf("error getting secret: %w", err)
	}

	return &model.Secret{
		Name:        name,
		Description: description,
		BankCard: &model.BankCard{
			CardNumber:  cardNumber,
			ExpiryMonth: expiryMonth,
			ExpiryYear:  expiryYear,
			CVV:         CVV,
		},
	}, nil
}

func (r *Postgres) GetSecretText(ctx context.Context, user, secretName string) (*model.Secret, error) {
	var name, description, text string

	query := `
            SELECT name, description, text
            FROM text_secrets
            WHERE username = $1
        `
	err := r.db.QueryRowContext(ctx, query, user).Scan(
		&name,
		&description,
		&text,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return &model.Secret{}, fmt.Errorf("Secret not found: %s", secretName)
	}
	if err != nil {
		return &model.Secret{}, fmt.Errorf("error getting secret: %w", err)
	}
	return &model.Secret{
		Name:        name,
		Description: description,
		SecretText: &model.SecretText{
			Text: text,
		},
	}, nil
}

func (r *Postgres) GetLoginPass(ctx context.Context, user, secretName string) (*model.Secret, error) {

	var name, description, login, password string

	query := `
            SELECT name, description, login, password
            FROM login_pass_secrets
            WHERE username = $1
        `

	err := r.db.QueryRowContext(ctx, query, user).Scan(
		&name,
		&description,
		&login,
		&password,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return &model.Secret{}, fmt.Errorf("Secret not found: %s", secretName)
	}
	if err != nil {
		return &model.Secret{}, fmt.Errorf("error getting secret: %w", err)
	}

	return &model.Secret{
		Name:        name,
		Description: description,
		LoginPass: &model.LoginPass{
			Login:    login,
			Password: password,
		},
	}, nil

}

func (r *Postgres) GetBinaryData(ctx context.Context, user, secretName string) (*model.Secret, error) {
	var name, description, data string

	query := `
            SELECT name, description, data
            FROM binary_secrets
            WHERE username = $1
        `
	err := r.db.QueryRowContext(ctx, query, user).Scan(
		&name,
		&description,
		&data,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return &model.Secret{}, fmt.Errorf("Secret not found: %s", secretName)
	}
	if err != nil {
		return &model.Secret{}, fmt.Errorf("error getting secret: %w", err)
	}
	return &model.Secret{
		Name:        name,
		Description: description,
		BinaryData: &model.BinaryData{
			Data: []byte(data),
		},
	}, nil
}
