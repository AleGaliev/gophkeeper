package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// Генерация RSA ключей
func GenerateRSAKeys(bits int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// Сохранение приватного ключа в файл
func SavePrivateKeyToFile(privateKey *rsa.PrivateKey, filename string) error {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	return os.WriteFile(filename, privateKeyPEM, 0600)
}

// Загрузка приватного ключа из файла
func LoadPrivateKeyFromFile(filename string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// Шифрование данных с использованием публичного ключа
func EncryptWithPublicKey(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, data, nil)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// Расшифровка данных с использованием приватного ключа
func DecryptWithPrivateKey(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func EncryptString(publicKey *rsa.PublicKey, text string) (string, error) {
	// Конвертируем строку в байты
	data := []byte(text)

	// Используем SHA256 для хэширования
	hash := sha256.New()

	// Шифруем с использованием OAEP
	ciphertext, err := rsa.EncryptOAEP(
		hash,
		rand.Reader,
		publicKey,
		data,
		nil, // label (дополнительные данные)
	)
	if err != nil {
		return "", fmt.Errorf("ошибка шифрования: %w", err)
	}

	// Кодируем в base64 для удобства хранения/передачи
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptString(privateKey *rsa.PrivateKey, encryptedText string) (string, error) {
	// Декодируем из base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования base64: %w", err)
	}

	// Используем SHA256 для хэширования
	hash := sha256.New()

	// Расшифровываем
	plaintext, err := rsa.DecryptOAEP(
		hash,
		rand.Reader,
		privateKey,
		ciphertext,
		nil, // label (дополнительные данные)
	)
	if err != nil {
		return "", fmt.Errorf("ошибка расшифровки: %w", err)
	}

	return string(plaintext), nil
}
