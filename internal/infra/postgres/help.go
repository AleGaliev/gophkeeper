package postgres

import "fmt"

func typeSecretInTableName(typeSecret string) (string, error) {
	var tableName string
	switch typeSecret {
	case "login_pass":
		tableName = "login_pass_secrets"
	case "card":
		tableName = "bank_card_secrets"
	case "text":
		tableName = "text_secrets"
	case "binary":
		tableName = "binary_secrets"
	default:
		return tableName, fmt.Errorf("unknown secret type: %s", typeSecret)
	}
	return tableName, nil
}
