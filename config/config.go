// internal/config/db.go

package config

import (
	"context"
	"fmt"
	"os"

	"fund-manager/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func LoadEnv() error {
	if _, err := os.Stat(".env"); err == nil {
		return godotenv.Load()
	}
	fmt.Println(".env file not found. Using system environment.")
	return nil
}

func InitDatabase(ctx context.Context) (*pgxpool.Pool, *repository.Queries, error) {
	connStr := os.Getenv("POSTGRES")
	if connStr == "" {
		return nil, nil, fmt.Errorf("POSTGRES env var not set")
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	return pool, repository.New(pool), nil
}
