package postgres

import (
	"fmt"
	"sso/internal/config"
)

func BuildDSN(cfg config.Storage) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
}
