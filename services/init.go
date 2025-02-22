package services

import (
	"database/sql"
	"fund-manager/internal/repository"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB
var Queries *repository.Queries

func ConnectToDb() {
	LoadEnv()
	connStr := os.Getenv("POSTGRES")
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}

	DB = db
	Queries = repository.New(DB)
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
