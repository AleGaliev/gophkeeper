package service

import (
	"context"
	"gophkeeper/internal/dto/model"
)

type Service interface {
	CreateUser(ctx context.Context, user model.User) (string, error)
	AuthUser(ctx context.Context, user model.User) (string, error)
	CreateSecret(ctx context.Context, user string, secret model.Secret) error
	UpdateSecret(ctx context.Context, user string, secret model.Secret) error
	DeleteSecret(ctx context.Context, user, secretType, secretName string) error
	GetSecretList(ctx context.Context, user, secretType string) ([]model.Secret, error)
	GetSecret(ctx context.Context, user, secretType, secretName string) (model.Secret, error)
	GetLoginFromToken(tokenString string) (string, error)
}
