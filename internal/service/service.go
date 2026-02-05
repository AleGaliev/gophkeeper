package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gophkeeper/internal/dto/model"
	"gophkeeper/internal/service/passhash"
)

const (
	ErrSecretExist = "this secret exists in storage"
)

type Storage interface {
	СreateLoginPass(ctx context.Context, user string, secret *model.Secret) error
	СreateSecretText(ctx context.Context, user string, secret *model.Secret) error
	CreateBinaryData(ctx context.Context, user string, secret *model.Secret) error
	CreateBankCard(ctx context.Context, user string, secret *model.Secret) error
	GetLoginPass(ctx context.Context, user, secretName string) (*model.Secret, error)
	GetBankCard(ctx context.Context, user, secretName string) (*model.Secret, error)
	GetBinaryData(ctx context.Context, user, secretName string) (*model.Secret, error)
	GetSecretText(ctx context.Context, user, secretName string) (*model.Secret, error)
	SecretExists(user, secretName, typeSecret string) (bool, error)
	UserExists(login string) (bool, error)
	CreateUser(user model.User) error
	GetUserHash(login string) (string, error)
	GetSecretList(user string, typeSecret string) (*model.Secrets, error)
}

type JWTManager interface {
	IssueJWT(user string) (string, error)
	GetLoginFromToken(tokenString string) (string, error)
}

type Service struct {
	storage    Storage
	JWTManager JWTManager
}

func New(storage Storage, JWTManager JWTManager) *Service {
	return &Service{
		storage:    storage,
		JWTManager: JWTManager,
	}
}

func (s *Service) CreateLoginPass(ctx context.Context, user string, message []byte) error {
	var secret model.Secret

	if err := json.Unmarshal(message, &secret); err != nil {
		return fmt.Errorf("%w", err)
	}

	check, err := s.storage.SecretExists(user, secret.Name, "login_pass")
	if err != nil {
		return err
	}
	if check {
		return errors.New(ErrSecretExist)
	}

	if err = s.storage.СreateLoginPass(ctx, user, &secret); err != nil {
		return err
	}
	return nil
}

func (s *Service) СreateSecretText(ctx context.Context, user string, message []byte) error {
	var secret model.Secret

	if err := json.Unmarshal(message, &secret); err != nil {
		return fmt.Errorf("%w", err)
	}

	check, err := s.storage.SecretExists(user, secret.Name, "text")
	if err != nil {
		return err
	}
	if check {
		return errors.New(ErrSecretExist)
	}
	if err = s.storage.СreateSecretText(ctx, user, &secret); err != nil {
		return err
	}
	return nil
}

func (s *Service) CreateBankCard(ctx context.Context, user string, message []byte) error {
	var secret model.Secret

	if err := json.Unmarshal(message, &secret); err != nil {
		return fmt.Errorf("%w", err)
	}

	check, err := s.storage.SecretExists(user, secret.Name, "card")
	if err != nil {
		return err
	}
	if check {
		return errors.New(ErrSecretExist)
	}
	if err = s.storage.CreateBankCard(ctx, user, &secret); err != nil {
		return err
	}
	return nil
}

func (s *Service) CreateBinaryData(ctx context.Context, user string, message []byte) error {
	var secret model.Secret

	if err := json.Unmarshal(message, &secret); err != nil {
		return fmt.Errorf("%w", err)
	}

	check, err := s.storage.SecretExists(user, secret.Name, "binary")
	if err != nil {
		return err
	}
	if check {
		return errors.New(ErrSecretExist)
	}
	if err = s.storage.CreateBinaryData(ctx, user, &secret); err != nil {
		return err
	}
	return nil
}

func (s *Service) CreateUser(message []byte) (string, error) {
	var (
		err  error
		user model.User
	)

	if err = json.Unmarshal(message, &user); err != nil {
		return "", fmt.Errorf("%w", err)
	}

	user.Hash, err = passhash.HashPassword(user.Password)
	if err != nil {
		return "", err
	}

	check, err := s.storage.UserExists(user.Login)
	if err != nil {
		return "", err
	}

	if check {
		return "", fmt.Errorf("user %s already exists", user.Login)
	}

	if err = s.storage.CreateUser(user); err != nil {
		return "", err
	}

	token, err := s.JWTManager.IssueJWT(user.Login)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) AuthUser(message []byte) (string, error) {
	var (
		err  error
		user model.User
	)

	if err = json.Unmarshal(message, &user); err != nil {
		return "", fmt.Errorf("%w", err)
	}

	user.Hash, err = passhash.HashPassword(user.Password)
	if err != nil {
		return "", err
	}

	hash, err := s.storage.GetUserHash(user.Login)
	if err != nil {
		return "", err
	}

	if !passhash.CheckPasswordHash(user.Password, hash) {
		return "", errors.New("invalid user/password")
	}

	token, err := s.JWTManager.IssueJWT(user.Login)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) GetSecretList(user string) ([]byte, error) {
	typeSecret := []string{
		"login_pass",
		"card",
		"text",
		"binary",
	}
	secrets := []model.Secrets{}
	for _, v := range typeSecret {
		secret, err := s.storage.GetSecretList(user, v)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, *secret)
	}
	return json.Marshal(secrets)
}

func (s *Service) GetSecret(ctx context.Context, user, secretType, secretName string) ([]byte, error) {
	var (
		secret *model.Secret
		err    error
	)
	switch secretType {
	case "text":
		secret, err = s.storage.GetSecretText(ctx, user, secretName)
	case "binary":
		secret, err = s.storage.GetBinaryData(ctx, user, secretName)
	case "login_pass":
		secret, err = s.storage.GetLoginPass(ctx, user, secretName)
	case "card":
		secret, err = s.storage.GetBankCard(ctx, user, secretName)
	default:
		return nil, fmt.Errorf("unknown secret type: %s", secretType)
	}

	if err != nil {
		return nil, err
	}

	return json.Marshal(secret)
}

func (s *Service) GetLoginFromToken(tokenString string) (string, error) {
	return s.JWTManager.GetLoginFromToken(tokenString)
}
