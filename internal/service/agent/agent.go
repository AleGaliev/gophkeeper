package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"gophkeeper/internal/config/agent"
	"gophkeeper/internal/dto/model"
	"gophkeeper/internal/infra/sender/grpc"
	"gophkeeper/internal/service/crypto/rsa"
	"strings"
	"time"
)

type Secret struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	SecretType  string      `json:"secret_type"`
	LoginPass   *LoginPass  `json:"login_pass,omitempty"`
	SecretText  *SecretText `json:"secret_text,omitempty"`
	BinaryData  *BinaryData `json:"binary_data,omitempty"`
	BankCard    *BankCard   `json:"bank_card,omitempty"`
}

type LoginPass struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type SecretText struct {
	Text string `json:"text"`
}

type BinaryData struct {
	Data []byte `json:"data"`
}

type BankCard struct {
	CardNumber  string `json:"card_number"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear  string `json:"expiry_year"`
	CVV         string `json:"cvv"`
}

type crypto interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

type sender interface {
	Register(ctx context.Context, user model.User) error
	Auth(ctx context.Context, user model.User) error
	CreateSecret(ctx context.Context, secret model.Secret) error
	UpdateSecret(ctx context.Context, secret model.Secret) error
	GetSecret(ctx context.Context, secretName string, secretType string) (*model.Secret, error)
	ListSecrets(ctx context.Context, secretType string) ([]*model.Secret, error)
	DeleteSecret(ctx context.Context, secretName string, secretType string) error
}

type Agent struct {
	sender sender
	crypto crypto
}

func New(cfg agent.Config) (*Agent, error) {
	sendler, err := grpc.New(cfg.GrpcAddr)
	if err != nil {
		return nil, err
	}
	crip, err := rsa.LoadPrivateKeyFromFile(cfg.KeyPath)
	if err != nil {
		fmt.Printf("Privat key not load, create new key in path %s (y/n): ", cfg.KeyPath)
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes" {
			rsaKey, err := rsa.GenerateRSAKeys()
			if err != nil {
				return nil, err
			}
			if err = rsaKey.SavePrivateKeyToFile(cfg.KeyPath); err != nil {
				return nil, err
			}
			fmt.Printf("RSA key saved to %s", cfg.KeyPath)

		} else {
			return nil, err
		}
	}

	return &Agent{
		sender: sendler,
		crypto: crip,
	}, nil
}

func (a *Agent) Register(ctx context.Context, user model.User) error {
	if err := a.sender.Register(ctx, user); err != nil {
		return err
	}
	return nil
}

func (a *Agent) Login(ctx context.Context, user model.User) error {
	if err := a.sender.Auth(ctx, user); err != nil {
		return err
	}
	return nil
}

func (a *Agent) createOrUpdateSecret(ctx context.Context, secret model.Secret, data []byte, operation func(context.Context, model.Secret) error) error {

	var err error
	if secret.SecretType == "binary_data" {
		secret.Data = data
		return operation(ctx, secret)
	}
	secret.Data, err = a.crypto.Encrypt(data)
	if err != nil {
		return err
	}
	return operation(ctx, secret)
}

func (a *Agent) CreateSecret(ctx context.Context, secret model.Secret, data []byte) error {
	return a.createOrUpdateSecret(ctx, secret, data, a.sender.CreateSecret)
}

func (a *Agent) UpdateSecret(ctx context.Context, secret model.Secret, data []byte) error {
	return a.createOrUpdateSecret(ctx, secret, data, a.sender.UpdateSecret)
}

func (a *Agent) GetSecret(ctx context.Context, secretName string, secretType string) (Secret, error) {
	secret, err := a.sender.GetSecret(ctx, secretName, secretType)
	if err != nil {
		fmt.Println(err.Error())
		time.Sleep(10 * time.Second)
		return Secret{}, err
	}

	return a.Decrypt(*secret)
}

func (a *Agent) ListSecrets(ctx context.Context, secretType string) ([]Secret, error) {
	secrets, err := a.sender.ListSecrets(ctx, secretType)
	if err != nil {
		return nil, err
	}
	result := make([]Secret, 0, len(secrets))
	for _, s := range secrets {
		secret := Secret{
			Name:        s.Name,
			Description: s.Description,
			SecretType:  s.SecretType,
		}

		result = append(result, secret)
	}

	return result, nil
}

func (a *Agent) DeleteSecret(ctx context.Context, secretName string, secretType string) error {
	if err := a.sender.DeleteSecret(ctx, secretName, secretType); err != nil {
		return err
	}
	return nil
}

//func (s Secret) Validate() error {
//	switch s.SecretType {
//	case "password":
//		if s.LoginPass.Login == "" || s.LoginPass.Password == "" {
//			return fmt.Errorf("login or password secret is empty")
//		}
//		if s.SecretText != nil || s.BinaryData != nil || s.BankCard != nil {
//			return fmt.Errorf("format secret not correct in type %s", s.SecretType)
//		}
//	case "text":
//		if s.SecretText.Text == "" {
//			return fmt.Errorf("text secret is empty")
//		}
//		if s.LoginPass != nil || s.BinaryData != nil || s.BankCard != nil {
//			return fmt.Errorf("format secret not correct in type %s", s.SecretType)
//		}
//	case "bank_card":
//		if s.BankCard.CVV == "" || s.BankCard.CardNumber == "" || s.BankCard.ExpiryMonth == "" || s.BankCard.ExpiryYear == "" {
//			return fmt.Errorf("bank_card secret is empty")
//		}
//		if s.LoginPass != nil || s.SecretText != nil || s.BankCard != nil {
//			return fmt.Errorf("format secret not correct in type %s", s.SecretType)
//		}
//	case "data":
//		if s.BinaryData == nil {
//			return fmt.Errorf("data secret is empty")
//		}
//		if s.LoginPass != nil || s.SecretText != nil || s.BankCard != nil {
//			return fmt.Errorf("format secret not correct in type %s", s.SecretType)
//		}
//	default:
//		return fmt.Errorf("unknown secret type %s list of available secret types [password, text, bank_card, data]", s.SecretType)
//	}
//	return nil
//}

func (a *Agent) Encrypt(secret Secret) ([]byte, error) {
	var (
		data []byte
		err  error
	)
	switch secret.SecretType {
	case "password":
		data, err = json.Marshal(secret.LoginPass)
		if err != nil {
			return nil, err
		}
	case "text":
		data, err = json.Marshal(secret.SecretText)
		if err != nil {
			return nil, err
		}
	case "bank_card":
		data, err = json.Marshal(secret.BankCard)
		if err != nil {
			return nil, err
		}
	case "data":
		return secret.BinaryData.Data, nil
	default:
		return nil, fmt.Errorf("unknown secret type %s of type %s", secret.SecretType, secret.SecretType)
	}

	return a.crypto.Encrypt(data)
}

func (a *Agent) Decrypt(sec model.Secret) (Secret, error) {
	var (
		decryptedData []byte
		err           error
	)
	if sec.SecretType != "binary_data" {
		decryptedData, err = a.crypto.Decrypt(sec.Data)
		if err != nil {
			return Secret{}, err
		}
	}

	secret := Secret{
		Name:        sec.Name,
		Description: sec.Description,
		SecretType:  sec.SecretType,
	}

	switch sec.SecretType {
	case "login_pass":
		if err := json.Unmarshal(decryptedData, &secret.LoginPass); err != nil {
			return secret, err
		}

	case "text":
		if err := json.Unmarshal(decryptedData, &secret.SecretText); err != nil {
			return secret, err
		}
	case "bank_card":
		if err := json.Unmarshal(decryptedData, &secret.BankCard); err != nil {
			return secret, err
		}
	case "binary_data":
		secret.BinaryData.Data = sec.Data
	default:
		return Secret{}, fmt.Errorf("unknown secret type %s", secret.SecretType)
	}

	return secret, nil
}
