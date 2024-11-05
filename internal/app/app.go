package app

import (
	"log/slog"
	grpcApp "sso/internal/app/grpc"
	"sso/internal/services/auth"
	"sso/internal/storage/postgres"
	"time"
)

type App struct {
	GRPCserver *grpcApp.App
	PGstorage  *postgres.Storage
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	storage, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, tokenTTL)
	grpcA := grpcApp.New(log, authService, grpcPort)
	return &App{GRPCserver: grpcA}
}
