package app

import (
	"context"
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	postgresapp "sso/internal/app/postgres"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
	Pg      *postgresapp.App
}

func New(log *slog.Logger, grpcPort int, tokenTTL time.Duration, dsn string) *App {
	// TODO: init storage
	pg := postgresapp.New(log, dsn)
	if err := pg.Run(context.Background()); err != nil {
		panic(err)
	}

	// TODO: init auth service

	grpcApp := grpcapp.New(log, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
