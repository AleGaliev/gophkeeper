package agent

import (
	"gophkeeper/internal/dto/model"
)

type Config struct {
	GrpcAddr string
	User     model.User
	KeyPath  string
}
