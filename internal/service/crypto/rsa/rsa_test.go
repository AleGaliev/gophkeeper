package rsa

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateRSAKeys(t *testing.T) {
	crypto, err := GenerateRSAKeys()
	if err != nil {
		t.Fatalf("Failed to generate RSA keys: %v", err)
	}

	if crypto.privateKey == nil {
		t.Fatal("Generated private key is nil")
	}

	if crypto.privateKey.N.BitLen() < 2047 {
		t.Errorf("Private key size is too small: %d bits", crypto.privateKey.N.BitLen())
	}
}

func TestSaveAndLoadPrivateKey(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private_key.pem")

	// Генерация и сохранение ключа
	crypto, err := GenerateRSAKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	err = crypto.SavePrivateKeyToFile(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to save private key: %v", err)
	}

	// Проверка, что файл создан
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		t.Fatalf("Private key file was not created: %v", err)
	}

	// Загрузка ключа
	loadedCrypto, err := LoadPrivateKeyFromFile(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to load private key: %v", err)
	}

	// Сравнение оригинального и загруженного ключа
	if loadedCrypto.privateKey.D.Cmp(crypto.privateKey.D) != 0 {
		t.Error("Loaded private key does not match original")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	data := []byte("Hello, world!")

	crypto, err := GenerateRSAKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	// Шифрование
	ciphertext, err := crypto.Encrypt(data)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Расшифровка
	plaintext, err := crypto.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Проверка соответствия данных
	if string(plaintext) != string(data) {
		t.Errorf("Decrypted data does not match original: expected %s, got %s", data, plaintext)
	}
}

func TestCreateCrypto(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "key.pem")

	crypto, err := CreateCrypto(privateKeyPath)
	if err != nil {
		t.Fatalf("CreateCrypto failed: %v", err)
	}

	if crypto.privateKey == nil {
		t.Fatal("CreateCrypto returned nil private key")
	}

	// Проверка, что файл ключа создан
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		t.Fatalf("CreateCrypto did not save private key file")
	}
}

// Тест с поддержкой пути ~/ (домашняя директория)
func TestSavePrivateKeyWithHomeDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get user home directory, skipping test")
	}

	tempSubDir := "test_rsa_keys"
	testPath := "~/" + tempSubDir + "/private_key.pem"
	fullExpectedPath := filepath.Join(home, tempSubDir, "private_key.pem")

	// Убедимся, что директория будет удалена
	defer os.RemoveAll(filepath.Join(home, tempSubDir))

	crypto, err := GenerateRSAKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	err = crypto.SavePrivateKeyToFile(testPath)
	if err != nil {
		t.Fatalf("Failed to save key with ~/ path: %v", err)
	}

	if _, err := os.Stat(fullExpectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected key file at %s but not found", fullExpectedPath)
	}
}
