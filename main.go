package main

import (
	"context"
	"database/sql"
	"fmt"
	"fund-manager/internal/repository"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	LoadEnv()
	connStr := os.Getenv("POSTGRES")
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	ctx := context.Background()
	queries := repository.New(db)

	stockList, err := queries.GetStocks(ctx)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(stockList)
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
