package grpc

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"net"
	auth1 "sso/internal/grpc/auth"
	auth2 "sso/internal/grpc/auth"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, auth auth2.Auth, port int) *App {
	recoveryOpst := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			log.Error("recovered from panic", slog.Any("panic", p))
			return status.Error(codes.Internal, "internal server error")
		}),
	}
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}
	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(recovery.UnaryServerInterceptor(recoveryOpst...), logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...)))
	auth1.Register(gRPCServer, auth)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}
func (a *App) Run() error {
	const op = "grpc.App.Run"
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	a.log.Info("starting grpc server", slog.String("address", l.Addr().String()))
	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpc.App.Stop"
	a.log.With(slog.String("op", op)).Info("stopping grpc server", slog.Int("port", a.port))
	a.gRPCServer.GracefulStop()
}
