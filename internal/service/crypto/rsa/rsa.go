package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
)

const bits = 2048

type Crypto struct {
	privateKey *rsa.PrivateKey
}

func CreateCrypto(privateKeyPath string) (*Crypto, error) {
	key, err := GenerateRSAKeys()
	if err != nil {
		return nil, err
	}
	if err = key.SavePrivateKeyToFile(privateKeyPath); err != nil {
		return nil, err
	}
	return key, nil
}

func GenerateRSAKeys() (*Crypto, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	return &Crypto{privateKey: privateKey}, nil
}

func (c *Crypto) SavePrivateKeyToFile(filename string) error {
	if filename[:2] == "~/" {
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

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(c.privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	return os.WriteFile(filename, privateKeyPEM, 0600)
}

func LoadPrivateKeyFromFile(filename string) (*Crypto, error) {
	if filename[:2] == "~/" {
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

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &Crypto{privateKey: privateKey}, nil
}

func (c *Crypto) Encrypt(data []byte) ([]byte, error) {
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, &c.privateKey.PublicKey, data, nil)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

func (c *Crypto) Decrypt(ciphertext []byte) ([]byte, error) {
	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, c.privateKey, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
