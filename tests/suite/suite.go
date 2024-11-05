package suite

import (
	"context"
	ssov1 "github.com/kerrek8/protos_sso1/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"sso/internal/config"
	"strconv"
	"testing"
)

type suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

const (
	grpcHost = "localhost"
)

func New(t *testing.T) (context.Context, *suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadPath("../config/local_tests.yaml")
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})
	grpcAddress := net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))

	cc, err := grpc.DialContext(context.Background(), grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {

		t.Fatalf("failed to dial grpc server: %v", err)
	}

	authClient := ssov1.NewAuthClient(cc)

	return ctx, &suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: authClient,
	}
}
