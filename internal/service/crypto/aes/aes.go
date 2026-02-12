package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/pem"
	"errors"
	"io"
	"os"
	"path/filepath"
)

const keySize = 32 // 32 bytes = 256 bits for AES-256

type Crypto struct {
	key []byte
}

func CreateCrypto(privateKeyPath string) (*Crypto, error) {
	crypto, err := GenerateAESKey()
	if err != nil {
		return nil, err
	}
	if err = crypto.SaveKeyToFile(privateKeyPath); err != nil {
		return nil, err
	}
	return crypto, nil
}

func GenerateAESKey() (*Crypto, error) {
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return &Crypto{key: key}, nil
}

func (c *Crypto) SaveKeyToFile(filename string) error {
	if len(filename) >= 2 && filename[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		filename = filepath.Join(homeDir, filename[2:])
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "AES KEY",
		Bytes: c.key,
	})
	return os.WriteFile(filename, keyPEM, 0600)
}

func LoadKeyFromFile(filename string) (*Crypto, error) {
	if len(filename) >= 2 && filename[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		filename = filepath.Join(homeDir, filename[2:])
	}
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	if len(block.Bytes) != keySize {
		return nil, errors.New("invalid key size")
	}

	return &Crypto{key: block.Bytes}, nil
}

func (c *Crypto) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and authenticate data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func (c *Crypto) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce and actual ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt and verify data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Additional helper function to create key from password
func CreateKeyFromPassword(password string) (*Crypto, error) {
	hash := sha256.Sum256([]byte(password))
	return &Crypto{key: hash[:]}, nil
}
