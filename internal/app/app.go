package app

import (
	"context"
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	postgresapp "sso/internal/app/postgres"
	"sso/internal/services/auth"
	"time"
)

type App struct {
	GRPCSrv  *grpcapp.App
	Postgres *postgresapp.App
}

func New(log *slog.Logger, grpcPort int, tokenTTL time.Duration, dsn string) *App {
	// TODO: init storage
	pgApp := postgresapp.New(log, dsn)
	if err := pgApp.Run(context.Background()); err != nil {
		panic(err)
	}

	// TODO: init auth service
	storage := pgApp.Storage()
	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv:  grpcApp,
		Postgres: pgApp,
	}
}
