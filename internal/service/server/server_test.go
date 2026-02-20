package server

import (
	"context"
	"errors"
	"gophkeeper/internal/dto/model"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockStorage(ctrl)
	mockJWT := NewMockJWTManager(ctrl)
	service := New(mockStorage, mockJWT)

	ctx := context.Background()
	testUser := model.User{
		Login:    "testuser",
		Password: "password123",
	}

	tests := []struct {
		name          string
		user          model.User
		mockSetup     func()
		expectedToken string
		expectedError string
	}{
		{
			name: "success",
			user: testUser,
			mockSetup: func() {
				mockStorage.EXPECT().UserExists(ctx, "testuser").Return(false, nil)
				mockStorage.EXPECT().CreateUser(ctx, gomock.Any()).Return(nil)
				mockJWT.EXPECT().IssueJWT("testuser").Return("test-token", nil)
			},
			expectedToken: "test-token",
			expectedError: "",
		},
		{
			name: "user already exists",
			user: testUser,
			mockSetup: func() {
				mockStorage.EXPECT().UserExists(ctx, "testuser").Return(true, nil)
			},
			expectedToken: "",
			expectedError: "user testuser already exists",
		},
		{
			name: "user exists check error",
			user: testUser,
			mockSetup: func() {
				mockStorage.EXPECT().UserExists(ctx, "testuser").Return(false, errors.New("db error"))
			},
			expectedToken: "",
			expectedError: "db error",
		},
		{
			name: "create user error",
			user: testUser,
			mockSetup: func() {
				mockStorage.EXPECT().UserExists(ctx, "testuser").Return(false, nil)
				mockStorage.EXPECT().CreateUser(ctx, gomock.Any()).Return(errors.New("create error"))
			},
			expectedToken: "",
			expectedError: "create error",
		},
		{
			name: "jwt issue error",
			user: testUser,
			mockSetup: func() {
				mockStorage.EXPECT().UserExists(ctx, "testuser").Return(false, nil)
				mockStorage.EXPECT().CreateUser(ctx, gomock.Any()).Return(nil)
				mockJWT.EXPECT().IssueJWT("testuser").Return("", errors.New("jwt error"))
			},
			expectedToken: "",
			expectedError: "jwt error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			token, err := service.CreateUser(ctx, tt.user)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestService_AuthUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockStorage(ctrl)
	mockJWT := NewMockJWTManager(ctrl)
	service := New(mockStorage, mockJWT)

	ctx := context.Background()
	testUser := model.User{
		Login:    "testuser",
		Password: "password123",
	}

	// Предварительно захешированный пароль (password123)
	hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrZ5WbYQYQYQYQYQYQYQYQYQYQYQYQYQYQ"

	tests := []struct {
		name          string
		user          model.User
		mockSetup     func()
		expectedToken string
		expectedError string
	}{
		{
			name: "invalid password",
			user: model.User{
				Login:    "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				mockStorage.EXPECT().GetUserHash(ctx, "testuser").Return(hashedPassword, nil)
			},
			expectedToken: "",
			expectedError: "invalid user/password",
		},
		{
			name: "user not found",
			user: testUser,
			mockSetup: func() {
				mockStorage.EXPECT().GetUserHash(ctx, "testuser").Return("", errors.New("user not found"))
			},
			expectedToken: "",
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			token, err := service.AuthUser(ctx, tt.user)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestService_CreateSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockStorage(ctrl)
	mockJWT := NewMockJWTManager(ctrl)
	service := New(mockStorage, mockJWT)

	ctx := context.Background()
	testSecret := model.Secret{
		Name:       "test",
		SecretType: "password",
		Data:       []byte("secret data"),
	}

	tests := []struct {
		name          string
		user          string
		secret        model.Secret
		mockSetup     func()
		expectedError string
	}{
		{
			name:   "success",
			user:   "testuser",
			secret: testSecret,
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(false, nil)
				mockStorage.EXPECT().СreateSecret(ctx, "testuser", &testSecret).Return(nil)
			},
			expectedError: "",
		},
		{
			name:   "secret already exists",
			user:   "testuser",
			secret: testSecret,
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(true, nil)
			},
			expectedError: ErrSecretExist,
		},
		{
			name:   "check exists error",
			user:   "testuser",
			secret: testSecret,
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(false, errors.New("db error"))
			},
			expectedError: "db error",
		},
		{
			name:   "create error",
			user:   "testuser",
			secret: testSecret,
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(false, nil)
				mockStorage.EXPECT().СreateSecret(ctx, "testuser", &testSecret).Return(errors.New("create error"))
			},
			expectedError: "create error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.CreateSecret(ctx, tt.user, tt.secret)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_UpdateSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockStorage(ctrl)
	mockJWT := NewMockJWTManager(ctrl)
	service := New(mockStorage, mockJWT)

	ctx := context.Background()
	testSecret := model.Secret{
		Name:       "test",
		SecretType: "password",
		Data:       []byte("updated data"),
	}

	tests := []struct {
		name          string
		user          string
		secret        model.Secret
		mockSetup     func()
		expectedError string
	}{
		{
			name:   "success",
			user:   "testuser",
			secret: testSecret,
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(true, nil)
				mockStorage.EXPECT().UpdateSecret(ctx, "testuser", &testSecret).Return(nil)
			},
			expectedError: "",
		},
		{
			name:   "secret not exists",
			user:   "testuser",
			secret: testSecret,
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(false, nil)
			},
			expectedError: ErrSecretNoExist,
		},
		{
			name:   "check exists error",
			user:   "testuser",
			secret: testSecret,
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(false, errors.New("db error"))
			},
			expectedError: "db error",
		},
		{
			name:   "update error",
			user:   "testuser",
			secret: testSecret,
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(true, nil)
				mockStorage.EXPECT().UpdateSecret(ctx, "testuser", &testSecret).Return(errors.New("update error"))
			},
			expectedError: "update error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.UpdateSecret(ctx, tt.user, tt.secret)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_DeleteSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockStorage(ctrl)
	mockJWT := NewMockJWTManager(ctrl)
	service := New(mockStorage, mockJWT)

	ctx := context.Background()

	tests := []struct {
		name          string
		user          string
		secretType    string
		secretName    string
		mockSetup     func()
		expectedError string
	}{
		{
			name:       "success",
			user:       "testuser",
			secretType: "password",
			secretName: "test",
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(true, nil)
				mockStorage.EXPECT().DeleteSecret(ctx, "testuser", "test", "password").Return(nil)
			},
			expectedError: "",
		},
		{
			name:       "secret not exists",
			user:       "testuser",
			secretType: "password",
			secretName: "test",
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(false, nil)
			},
			expectedError: "",
		},
		{
			name:       "check exists error",
			user:       "testuser",
			secretType: "password",
			secretName: "test",
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(false, errors.New("db error"))
			},
			expectedError: "db error",
		},
		{
			name:       "delete error",
			user:       "testuser",
			secretType: "password",
			secretName: "test",
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(true, nil)
				mockStorage.EXPECT().DeleteSecret(ctx, "testuser", "test", "password").Return(errors.New("delete error"))
			},
			expectedError: "delete error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.DeleteSecret(ctx, tt.user, tt.secretType, tt.secretName)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockStorage(ctrl)
	mockJWT := NewMockJWTManager(ctrl)
	service := New(mockStorage, mockJWT)

	ctx := context.Background()
	expectedSecret := model.Secret{
		Name:       "test",
		SecretType: "password",
		Data:       []byte("secret data"),
	}

	tests := []struct {
		name           string
		user           string
		secretType     string
		secretName     string
		mockSetup      func()
		expectedSecret model.Secret
		expectedError  string
	}{
		{
			name:       "success",
			user:       "testuser",
			secretType: "password",
			secretName: "test",
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(true, nil)
				mockStorage.EXPECT().GetSecret(ctx, "testuser", "test", "password").Return(expectedSecret, nil)
			},
			expectedSecret: expectedSecret,
			expectedError:  "",
		},
		{
			name:       "secret not exists",
			user:       "testuser",
			secretType: "password",
			secretName: "test",
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(false, nil)
			},
			expectedSecret: model.Secret{},
			expectedError:  ErrSecretNoExist,
		},
		{
			name:       "check exists error",
			user:       "testuser",
			secretType: "password",
			secretName: "test",
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(false, errors.New("db error"))
			},
			expectedSecret: model.Secret{},
			expectedError:  "db error",
		},
		{
			name:       "get error",
			user:       "testuser",
			secretType: "password",
			secretName: "test",
			mockSetup: func() {
				mockStorage.EXPECT().SecretExists(ctx, "testuser", "test", "password").Return(true, nil)
				mockStorage.EXPECT().GetSecret(ctx, "testuser", "test", "password").Return(model.Secret{}, errors.New("get error"))
			},
			expectedSecret: model.Secret{},
			expectedError:  "get error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			secret, err := service.GetSecret(ctx, tt.user, tt.secretType, tt.secretName)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, tt.expectedSecret, secret)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSecret, secret)
			}
		})
	}
}

func TestService_GetSecretList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockStorage(ctrl)
	mockJWT := NewMockJWTManager(ctrl)
	service := New(mockStorage, mockJWT)

	ctx := context.Background()
	expectedSecrets := []model.Secret{
		{Name: "secret1", SecretType: "password", Data: []byte("data1")},
		{Name: "secret2", SecretType: "card", Data: []byte("data2")},
	}

	tests := []struct {
		name            string
		user            string
		secretType      string
		mockSetup       func()
		expectedSecrets []model.Secret
		expectedError   string
	}{
		{
			name:       "success all secrets",
			user:       "testuser",
			secretType: "",
			mockSetup: func() {
				mockStorage.EXPECT().GetSecretList(ctx, "testuser").Return(expectedSecrets, nil)
			},
			expectedSecrets: expectedSecrets,
			expectedError:   "",
		},
		{
			name:       "success filtered by type",
			user:       "testuser",
			secretType: "password",
			mockSetup: func() {
				mockStorage.EXPECT().GetSecretListInType(ctx, "testuser", "password").
					Return([]model.Secret{expectedSecrets[0]}, nil)
			},
			expectedSecrets: []model.Secret{expectedSecrets[0]},
			expectedError:   "",
		},
		{
			name:       "error get all secrets",
			user:       "testuser",
			secretType: "",
			mockSetup: func() {
				mockStorage.EXPECT().GetSecretList(ctx, "testuser").Return(nil, errors.New("db error"))
			},
			expectedSecrets: nil,
			expectedError:   "db error",
		},
		{
			name:       "error get filtered secrets",
			user:       "testuser",
			secretType: "password",
			mockSetup: func() {
				mockStorage.EXPECT().GetSecretListInType(ctx, "testuser", "password").
					Return(nil, errors.New("db error"))
			},
			expectedSecrets: nil,
			expectedError:   "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			secrets, err := service.GetSecretList(ctx, tt.user, tt.secretType)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, secrets)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSecrets, secrets)
			}
		})
	}
}

func TestService_GetLoginFromToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockStorage(ctrl)
	mockJWT := NewMockJWTManager(ctrl)
	service := New(mockStorage, mockJWT)

	tests := []struct {
		name          string
		token         string
		mockSetup     func()
		expectedLogin string
		expectedError string
	}{
		{
			name:  "success",
			token: "valid-token",
			mockSetup: func() {
				mockJWT.EXPECT().GetLoginFromToken("valid-token").Return("testuser", nil)
			},
			expectedLogin: "testuser",
			expectedError: "",
		},
		{
			name:  "error",
			token: "invalid-token",
			mockSetup: func() {
				mockJWT.EXPECT().GetLoginFromToken("invalid-token").Return("", errors.New("invalid token"))
			},
			expectedLogin: "",
			expectedError: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			login, err := service.GetLoginFromToken(tt.token)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, login)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLogin, login)
			}
		})
	}
}
