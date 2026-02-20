package jwtmanager

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewJWTManager_WithProvidedKey(t *testing.T) {
	key := "my-secret-key-32-chars-long!!!!"
	manager := New(key)

	assert.NotNil(t, manager)
	assert.Equal(t, []byte(key), manager.key)
}

func TestNewJWTManager_WithoutKey_GeneratesRandom(t *testing.T) {
	manager := New("")

	assert.NotNil(t, manager)
	assert.Len(t, manager.key, 32) // 256 bits
}

func TestIssueJWT_ValidToken(t *testing.T) {
	manager := New("test-key-32-chars-long!!!!!!!!")
	tokenString, err := manager.IssueJWT("testuser")

	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Parse and validate manually
	parsedToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return manager.key, nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	assert.True(t, ok)
	assert.Equal(t, "testuser", claims.Subject)
	assert.Equal(t, issuer, claims.Issuer)
}

func TestValidateToken_Valid(t *testing.T) {
	manager := New("test-key-32-chars-long!!!!!!!!")
	tokenString, _ := manager.IssueJWT("testuser")

	claims, err := manager.ValidateToken(tokenString)

	assert.NoError(t, err)
	assert.Equal(t, "testuser", claims.Subject)
	assert.Equal(t, issuer, claims.Issuer)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	manager1 := New("key111111111111111111111111111111")
	manager2 := New("key222222222222222222222222222222") // другой ключ

	tokenString, _ := manager1.IssueJWT("testuser")
	_, err := manager2.ValidateToken(tokenString)

	assert.Error(t, err)
}

func TestValidateToken_Expired(t *testing.T) {
	key := "test-key-32-chars-long!!!!!!!!"
	manager := New(key)

	// Создаём просроченный токен вручную
	claims := jwt.RegisteredClaims{
		Subject:   "testuser",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // уже истёк
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		Issuer:    issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString([]byte(key))

	_, err := manager.ValidateToken(signedToken)
	assert.Error(t, err)
}

func TestGetLoginFromToken_Valid(t *testing.T) {
	manager := New("test-key-32-chars-long!!!!!!!!")
	tokenString, _ := manager.IssueJWT("alice")

	login, err := manager.GetLoginFromToken(tokenString)

	assert.NoError(t, err)
	assert.Equal(t, "alice", login)
}

func TestGetLoginFromToken_Invalid(t *testing.T) {
	manager := New("test-key-32-chars-long!!!!!!!!")
	login, err := manager.GetLoginFromToken("invalid.token.string")

	assert.Error(t, err)
	assert.Empty(t, login)
}
