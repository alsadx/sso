package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type App struct {
	log  *slog.Logger
	pool *pgxpool.Pool
	dsn  string
}

func New(log *slog.Logger, dsn string) *App {
	return &App{
		log: log,
		dsn: dsn,
	}
}

func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		panic(err)
	}
}

func (a *App) Run(ctx context.Context) error {
	const op = "postgresapp.Run"

	log := a.log.With(slog.String("op", op))

	cfg, err := pgxpool.ParseConfig(a.dsn)
	if err != nil {
		return fmt.Errorf("%s: parse config: %w", op, err)
	}

	//cfg.MaxConns = 10
	//cfg.MinConns = 2
	//cfg.MaxConnLifetime = time.Hour
	//cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("%s: connect: %w", op, err)
	}

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("%s: ping: %w", op, err)
	}

	a.pool = pool

	log.Info("connected to PostgreSQL")

	return nil
}

func (a *App) Stop() {
	const op = "postgresapp.Stop"

	log := a.log.With(slog.String("op", op))

	log.Info("closing PostgreSQL connection pool")

	// Закрывает пул и ждёт завершения активных запросов
	a.pool.Close()
}

func (a *App) Pool() *pgxpool.Pool {
	return a.pool
}
