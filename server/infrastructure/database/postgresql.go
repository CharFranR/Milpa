package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePool(ctx context.Context, DatabaseURL string) (*pgxpool.Pool, error) {

	config, err := pgxpool.ParseConfig(DatabaseURL)
	if err != nil {
		return nil, err
	}

	config.MinConns = 5
	config.MaxConns = 25
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 15 * time.Minute

	return pgxpool.NewWithConfig(ctx, config)

}

func CloseConnection(connection *pgxpool.Pool) {
	connection.Close()
}
