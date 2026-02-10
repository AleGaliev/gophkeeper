package model

const (
	TypeSecretLoginPassword = "login_pass"
	TypeSecretText          = "text"
	TypeSecretBankCard      = "bank_card"
	TypeSecretBinaryData    = "binary_data"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Hash     string `json:"hash"`
}

type Secrets struct {
	SecretType string   `json:"secret_type"`
	Secrets    []Secret `json:"secrets"`
}

type Secret struct {
	Name        string `json:"name"`
	SecretType  string `json:"secret_type"`
	Description string `json:"description"`
	Data        []byte `json:"data"`
}

//type Secret struct {
//	Name        string      `json:"name"`
//	Description string      `json:"description"`
//	LoginPass   *LoginPass  `json:"login_pass,omitempty"`
//	SecretText  *SecretText `json:"secret_text,omitempty"`
//	BinaryData  *BinaryData `json:"binary_data,omitempty"`
//	BankCard    *BankCard   `json:"bank_card,omitempty"`
//}

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
