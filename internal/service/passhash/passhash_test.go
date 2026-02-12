package passhash

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "mySecurePassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if hash == "" {
		t.Error("Expected hash to be non-empty")
	}

	// Проверяем, что хеш не совпадает с исходным паролем
	if hash == password {
		t.Error("Expected hash to be different from password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "mySecurePassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Проверяем, что правильный пароль проходит проверку
	if !CheckPasswordHash(password, hash) {
		t.Error("Expected password to match hash")
	}

	// Проверяем, что неправильный пароль не проходит проверку
	wrongPassword := "wrongPassword"
	if CheckPasswordHash(wrongPassword, hash) {
		t.Error("Expected wrong password to not match hash")
	}
}
