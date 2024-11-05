package auth

import (
	"context"
	"errors"
	ssov1 "github.com/kerrek8/protos_sso1/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso/internal/services/auth"
	"sso/internal/storage"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
}

// Register registers the AuthServer with the gRPC server.
func Register(gRPCServer *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	in *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if in.Email == "" || in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	if in.AppId == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}
	token, err := s.auth.Login(ctx, in.Email, in.Password, int(in.AppId))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}
	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	in *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if in.Email == "" || in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	uid, err := s.auth.RegisterNewUser(ctx, in.Email, in.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register")
	}
	return &ssov1.RegisterResponse{UserId: uid}, nil
}
