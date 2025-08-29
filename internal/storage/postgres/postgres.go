package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"sso/internal/domain/models"
	"sso/internal/storage"
)

type Storage struct {
	DbPool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{DbPool: pool}
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error) {
	const op = "storage.postgres.SaveUser"

	query := `INSERT INTO users (email, pass_hash) VALUES ($1, $2) RETURNING id`

	err = s.DbPool.QueryRow(ctx, query, email, passHash).Scan(&uid)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return
}

func (s *Storage) User(ctx context.Context, email string) (user models.User, err error) {
	const op = "storage.postgres.User"

	user = models.User{}

	query := `SELECT id, email, pass_hash FROM users WHERE email = $1`

	err = s.DbPool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return user, fmt.Errorf("%s: %w", op, err)
	}

	return
}

func (s *Storage) App(ctx context.Context, id int) (app models.App, err error) {
	const op = "storage.postgres.App"

	app = models.App{}

	query := `SELECT id, name, secret FROM users WHERE id = $1`

	err = s.DbPool.QueryRow(ctx, query, id).Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return app, fmt.Errorf("%s: %w", op, err)
	}

	return
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error) {
	const op = "storage.postgres.IsAdmin"

	query := `SELECT EXISTS (SELECT 1 FROM admins WHERE user_id = $1)`

	err = s.DbPool.QueryRow(ctx, query, userID).Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return isAdmin, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return false, err
	}

	return
}
