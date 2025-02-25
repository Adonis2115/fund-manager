package initializers

import (
	"context"
	"fund-manager/internal/repository"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var Pool *pgxpool.Pool
var Queries *repository.Queries

func ConnectToDb() {
	LoadEnv()
	connStr := os.Getenv("POSTGRES")
	pool, err := pgxpool.New(context.Background(), connStr)

	if err != nil {
		log.Fatal(err)
	}

	Pool = pool
	Queries = repository.New(Pool)
}

func LoadEnv() {
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file: ", err)
		}
	} else {
		log.Println(".env file not found. Loading environment variables from the system.")
	}
}
