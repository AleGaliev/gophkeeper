package grpc

import (
	"context"
	"fmt"
	"gophkeeper/internal/dto/model"
	pb "gophkeeper/internal/pkg/proto"
	"gophkeeper/internal/pkg/proto/mapper"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Grpc struct {
	client pb.GophKeeperClient
	token  string
	mu     sync.RWMutex
}

func New(addr string) (*Grpc, error) {
	// Создаем соединение с интерцептором
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//Разобраться почему не работает Interceptor
		//grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		//
		//	if g, ok := ctx.Value("grpcClient").(*Grpc); ok {
		//		g.mu.RLock()
		//		token := g.token
		//		g.mu.RUnlock()
		//
		//		if token != "" {
		//
		//			ctx = metadata.AppendToOutgoingContext(
		//				ctx,
		//				"authorization", token,
		//			)
		//		}
		//	}
		//	return invoker(ctx, method, req, reply, cc, opts...)
		//}),
	)
	if err != nil {
		return nil, err
	}

	grpcClient := pb.NewGophKeeperClient(conn)

	return &Grpc{
		client: grpcClient,
		token:  "",
	}, nil
}

func (g *Grpc) SetToken(token string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.token = token
}

func (g *Grpc) Register(ctx context.Context, user model.User) error {
	pbUser := mapper.UserInProtoUser(user)

	_, err := g.client.Register(ctx, pb.LoginRequest_builder{
		User: pbUser,
	}.Build())

	if err != nil {
		return err
	}

	var header metadata.MD
	token := ""
	if authHeaders := header.Get("authorization"); len(authHeaders) > 0 {
		token = authHeaders[0]
	} else {
		return fmt.Errorf("no authorization header")
	}

	g.mu.Lock()
	g.token = token
	g.mu.Unlock()

	return nil
}

func (g *Grpc) Auth(ctx context.Context, user model.User) error {
	pbUser := mapper.UserInProtoUser(user)
	var header metadata.MD

	_, err := g.client.Login(ctx, pb.LoginRequest_builder{
		User: pbUser,
	}.Build(), grpc.Header(&header))

	if err != nil {
		return err
	}
	token := ""
	if authHeaders := header.Get("authorization"); len(authHeaders) > 0 {
		token = authHeaders[0]
	} else {
		return fmt.Errorf("no authorization header")
	}

	g.mu.Lock()
	g.token = token
	g.mu.Unlock()

	return nil
}

func (g *Grpc) CreateSecret(ctx context.Context, secret model.Secret) error {
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		"authorization", g.token,
	)
	pbSecret := mapper.SecretToProto(secret)
	_, err := g.client.CreateSecret(ctx, pb.CreateSecretRequest_builder{
		Secret: pbSecret,
	}.Build())

	if err != nil {
		return err
	}

	return nil
}

func (g *Grpc) UpdateSecret(ctx context.Context, secret model.Secret) error {
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		"authorization", g.token,
	)
	pbSecret := mapper.SecretToProto(secret)

	_, err := g.client.UpdateSecret(ctx, pb.UpdateSecretRequest_builder{
		Secret: pbSecret,
	}.Build())

	if err != nil {
		return err
	}

	return nil
}

func (g *Grpc) GetSecret(ctx context.Context, secretName, secretType string) (*model.Secret, error) {
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		"authorization", g.token,
	)
	updateSecretResponse, err := g.client.GetSecret(ctx, pb.GetSecretRequest_builder{
		SecretName: secretName,
		SecretType: mapper.StringToProtoSecretType(secretType),
	}.Build())

	if err != nil {
		return nil, err
	}

	secret := mapper.ProtoSecretInSecret(updateSecretResponse.GetSecret())

	return secret, nil
}

func (g *Grpc) ListSecrets(ctx context.Context, secretType string) ([]*model.Secret, error) {
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		"authorization", g.token,
	)
	getListSecretResponse, err := g.client.ListSecrets(ctx, pb.GetListSecretsRequest_builder{
		SecretType: mapper.StringToProtoSecretType(secretType),
	}.Build())

	if err != nil {
		return nil, err
	}

	secrets := mapper.SliceProtoSecretToSliceSecret(getListSecretResponse.GetSecrets())

	return secrets, nil
}

func (g *Grpc) DeleteSecret(ctx context.Context, secretName, secretType string) error {
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		"authorization", g.token,
	)
	_, err := g.client.DeleteSecret(ctx, pb.DeleteSecretRequest_builder{
		SecretName: secretName,
		SecretType: mapper.StringToProtoSecretType(secretType),
	}.Build())

	if err != nil {
		return err
	}

	return nil
}
