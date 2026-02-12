package handlers

import (
	"context"
	"fmt"
	"gophkeeper/internal/dto/log"
	"gophkeeper/internal/dto/service"
	pb "gophkeeper/internal/pkg/proto"
	"gophkeeper/internal/pkg/proto/mapper"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)
}

type Handler struct {
	pb.UnimplementedGophKeeperServer
	service service.Service
	logger  log.Logger
}

func New(service service.Service, logger log.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Register(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	resp := &pb.LoginResponse{}
	h.logger.Debug(ctx, "info request", "metod", "register", "user", req.GetUser().GetLogin(), "less pass", len(req.GetUser().GetPassword()))

	token, err := h.service.CreateUser(ctx, *mapper.ProtoUserInUser(*req.GetUser()))
	if err != nil {
		return resp, err
	}
	md := metadata.Pairs(
		"Authorization", token,
	)

	if err = grpc.SetHeader(ctx, md); err != nil {
		h.logger.Error(ctx, "info request", "metod", "register", "error", err.Error())

		return resp, err
	}
	return resp, nil
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	resp := &pb.LoginResponse{}
	h.logger.Debug(ctx, "info request", "metod", "login", "user", req.GetUser().GetLogin(), "less pass", len(req.GetUser().GetPassword()))

	token, err := h.service.AuthUser(ctx, *mapper.ProtoUserInUser(*req.GetUser()))
	if err != nil {
		h.logger.Error(ctx, "info request", "metod", "login", "error", err.Error())

		return resp, err
	}
	md := metadata.Pairs(
		"authorization", token,
	)

	if err = grpc.SetHeader(ctx, md); err != nil {
		h.logger.Error(ctx, "info request", "metod", "login", "error", err.Error())
		return resp, err
	}
	return resp, nil
}

func (h *Handler) CreateSecret(ctx context.Context, req *pb.CreateSecretRequest) (*pb.CreateSecretResponse, error) {
	resp := &pb.CreateSecretResponse{}
	h.logger.Debug(ctx, "info request", "metod", "createSecret", "secret_name", req.GetSecret().GetName(), "secret_type", req.GetSecret().GetSecretType(), "len_data", len(req.GetSecret().GetEncryptedData()))
	secret := mapper.ProtoSecretInSecret(req.GetSecret())

	user, err := getUserFromContext(ctx)
	if err != nil {
		h.logger.Error(ctx, "info request", "metod", "createSecret", "error", err.Error())
		return resp, fmt.Errorf("user not found in context")
	}

	if err := h.service.CreateSecret(ctx, user, *secret); err != nil {
		h.logger.Error(ctx, "info request", "metod", "createSecret", "error", err.Error())
		return resp, err
	}

	return resp, nil
}

func (h *Handler) UpdateSecret(ctx context.Context, req *pb.UpdateSecretRequest) (*pb.UpdateSecretResponse, error) {
	resp := &pb.UpdateSecretResponse{}
	h.logger.Debug(ctx, "info request", "metod", "updateSecret", "secret_name", req.GetSecret().GetName(), "secret_type", req.GetSecret().GetSecretType(), "len_data", len(req.GetSecret().GetEncryptedData()))

	secret := mapper.ProtoSecretToSecret(*req.GetSecret())

	user, err := getUserFromContext(ctx)
	if err != nil {
		h.logger.Error(ctx, "info request", "metod", "updateSecret", "error", err.Error())

		return resp, fmt.Errorf("user not found in context")
	}

	if err := h.service.UpdateSecret(ctx, user, *secret); err != nil {
		h.logger.Error(ctx, "info request", "metod", "updateSecret", "error", err.Error())
		return resp, err
	}

	return resp, nil
}

func (h *Handler) DeleteSecret(ctx context.Context, req *pb.DeleteSecretRequest) (*pb.DeleteSecretResponse, error) {
	resp := &pb.DeleteSecretResponse{}
	h.logger.Debug(ctx, "info request", "metod", "deleteSecret", "secret_name", req.GetSecretName(), "secret_type", req.GetSecretType())

	user, err := getUserFromContext(ctx)
	if err != nil {
		h.logger.Error(ctx, "info request", "metod", "deleteSecret", "error", err.Error())
		return resp, err
	}

	SecretType := req.GetSecretType()
	SecretName := req.GetSecretName()

	if err := h.service.DeleteSecret(ctx, user, mapper.ProtoSecretTypeStringTo(SecretType), SecretName); err != nil {
		h.logger.Error(ctx, "info request", "metod", "deleteSecret", "error", err.Error())
		return resp, err
	}

	return resp, nil
}

func (h *Handler) GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	resp := &pb.GetSecretResponse{}
	h.logger.Debug(ctx, "info request", "metod", "getSecret", "secret_name", req.GetSecretName(), "secret_type", req.GetSecretType())

	user, err := getUserFromContext(ctx)
	if err != nil {
		h.logger.Error(ctx, "info request", "metod", "getSecret", "error", err.Error())
		return resp, err
	}

	SecretType := req.GetSecretType()
	SecretName := req.GetSecretName()

	secret, err := h.service.GetSecret(ctx, user, mapper.ProtoSecretTypeStringTo(SecretType), SecretName)
	if err != nil {
		h.logger.Error(ctx, "info request", "metod", "getSecret", "error", err.Error())
		return resp, err
	}

	resp.SetSecret(pb.Secret_builder{
		Name:          secret.Name,
		SecretType:    mapper.StringToProtoSecretType(secret.SecretType),
		Description:   secret.Description,
		EncryptedData: secret.Data,
	}.Build())

	return resp, nil
}

func (h *Handler) ListSecrets(ctx context.Context, req *pb.GetListSecretsRequest) (*pb.GetListSecretsResponse, error) {
	resp := &pb.GetListSecretsResponse{}
	h.logger.Debug(ctx, "info request", "metod", "listSecret")

	user, err := getUserFromContext(ctx)
	if err != nil {
		h.logger.Error(ctx, "info request", "metod", "listSecret", "error", err.Error())
		return resp, err
	}

	secrets, err := h.service.GetSecretList(ctx, user, mapper.ProtoSecretTypeStringTo(req.GetSecretType()))
	if err != nil {
		h.logger.Error(ctx, "info request", "metod", "listSecret", "error", err.Error())
		return resp, err
	}

	resp.SetSecrets(mapper.SecretSliceToProto(secrets))
	return resp, nil
}

func getUserFromContext(ctx context.Context) (string, error) {
	userValue := ctx.Value("user")
	user, ok := userValue.(string)
	if !ok {
		return "", fmt.Errorf("user not found in context")
	}
	return user, nil
}

func GetListNotAuthMetod() []string {
	return []string{
		"/gophkeeper.GophKeeper/Register",
		"/gophkeeper.GophKeeper/Login",
	}
}
