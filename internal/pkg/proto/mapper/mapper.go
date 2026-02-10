package mapper

import (
	"gophkeeper/internal/dto/model"
	pb "gophkeeper/internal/pkg/proto"
)

func ProtoUserInUser(user pb.User) *model.User {
	return &model.User{
		Login:    user.GetLogin(),
		Password: user.GetPassword(),
	}
}

func ProtoSecertInSecert(pbSecret *pb.Secret) *model.Secret {
	secret := model.Secret{
		Name:        pbSecret.GetName(),
		Description: pbSecret.GetDescription(),
		SecretType:  ProtoSecretTypeStringTo(pbSecret.GetSecretType()),
		Data:        pbSecret.GetEncryptedData(),
	}
	return &secret
}

func SecretSliceToProto(metrics []model.Secret) []*pb.Secret {
	result := make([]*pb.Secret, 0, len(metrics))
	for _, m := range metrics {
		result = append(result, SecretToProto(m))
	}
	return result
}

func SecretToProto(m model.Secret) *pb.Secret {
	return pb.Secret_builder{
		Name:          m.Name,
		Description:   m.Description,
		EncryptedData: m.Data,
	}.Build()
}

func StringToProtoSecretType(secretTypeStr string) pb.SecretType {
	// Простое преобразование - можно улучшить при необходимости
	switch secretTypeStr {
	case model.TypeSecretLoginPassword:
		return pb.SecretType_SECRET_TYPE_LOGIN_PASSWORD
	case model.TypeSecretText:
		return pb.SecretType_SECRET_TYPE_TEXT
	case model.TypeSecretBinaryData:
		return pb.SecretType_DSECRET_TYPE_BINARY
	case model.TypeSecretBankCard:
		return pb.SecretType_DSECRET_TYPE_CARD
	}
	return pb.SecretType_SECRET_TYPE_UNSPECIFIED
}

func ProtoSecretTypeStringTo(secretTypeProto pb.SecretType) string {
	// Простое преобразование - можно улучшить при необходимости
	switch secretTypeProto {
	case pb.SecretType_SECRET_TYPE_LOGIN_PASSWORD:
		return model.TypeSecretLoginPassword
	case pb.SecretType_SECRET_TYPE_TEXT:
		return model.TypeSecretText
	case pb.SecretType_DSECRET_TYPE_BINARY:
		return model.TypeSecretBinaryData
	case pb.SecretType_DSECRET_TYPE_CARD:
		return model.TypeSecretBankCard
	}
	return ""
}
