package server

import (
	"context"
	"errors"
	"fmt"
	"gophkeeper/internal/dto/model"
	"gophkeeper/internal/service/passhash"
)

const (
	ErrSecretExist   = "this secret exists in storage"
	ErrSecretNoExist = "this secret no exists in storage"
)

type Storage interface {
	СreateSecret(ctx context.Context, user string, secret *model.Secret) error
	UpdateSecret(ctx context.Context, user string, secret *model.Secret) error
	GetSecret(ctx context.Context, user, secretName, secretType string) (model.Secret, error)
	SecretExists(ctx context.Context, user, secretName, typeSecret string) (bool, error)
	UserExists(ctx context.Context, login string) (bool, error)
	CreateUser(ctx context.Context, user model.User) error
	GetUserHash(ctx context.Context, login string) (string, error)
	GetSecretList(ctx context.Context, user string) ([]model.Secret, error)
	GetSecretListInType(ctx context.Context, user, secretType string) ([]model.Secret, error)
	DeleteSecret(ctx context.Context, user, secretName, secretTYPE string) error
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

func (s *Service) CreateUser(ctx context.Context, user model.User) (string, error) {
	var err error

	user.Hash, err = passhash.HashPassword(user.Password)
	if err != nil {
		return "", err
	}
	check, err := s.storage.UserExists(ctx, user.Login)
	if err != nil {
		return "", err
	}

	if check {
		return "", fmt.Errorf("user %s already exists", user.Login)
	}

	if err = s.storage.CreateUser(ctx, user); err != nil {
		return "", err
	}

	token, err := s.JWTManager.IssueJWT(user.Login)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) AuthUser(ctx context.Context, user model.User) (string, error) {
	var err error
	user.Hash, err = passhash.HashPassword(user.Password)
	if err != nil {
		return "", err
	}

	hash, err := s.storage.GetUserHash(ctx, user.Login)
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

func (s *Service) CreateSecret(ctx context.Context, user string, secret model.Secret) error {
	check, err := s.storage.SecretExists(ctx, user, secret.Name, secret.SecretType)
	if err != nil {
		return err
	}
	if check {
		return errors.New(ErrSecretExist)
	}

	if err = s.storage.СreateSecret(ctx, user, &secret); err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateSecret(ctx context.Context, user string, secret model.Secret) error {
	check, err := s.storage.SecretExists(ctx, user, secret.Name, secret.SecretType)
	if err != nil {
		return err
	}
	if !check {
		return errors.New(ErrSecretNoExist)
	}

	if err = s.storage.UpdateSecret(ctx, user, &secret); err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteSecret(ctx context.Context, user, secretType, secretName string) error {
	check, err := s.storage.SecretExists(ctx, user, secretName, secretType)
	if err != nil {
		return err
	}
	if !check {
		return nil
	}

	if err = s.storage.DeleteSecret(ctx, user, secretName, secretType); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetSecret(ctx context.Context, user, secretType, secretName string) (model.Secret, error) {
	check, err := s.storage.SecretExists(ctx, user, secretName, secretType)
	if err != nil {
		return model.Secret{}, err
	}
	if !check {
		return model.Secret{}, errors.New(ErrSecretNoExist)
	}

	secret, err := s.storage.GetSecret(ctx, user, secretName, secretType)

	if err != nil {
		return model.Secret{}, err
	}
	return secret, nil
}

func (s *Service) GetSecretList(ctx context.Context, user, secretType string) ([]model.Secret, error) {
	var (
		secrets []model.Secret
		err     error
	)

	if secretType != "" {
		secrets, err = s.storage.GetSecretListInType(ctx, user, secretType)
		if err != nil {
			return nil, err
		}
		return secrets, nil
	}

	secret, err := s.storage.GetSecretList(ctx, user)

	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (s *Service) GetLoginFromToken(tokenString string) (string, error) {
	return s.JWTManager.GetLoginFromToken(tokenString)
}
