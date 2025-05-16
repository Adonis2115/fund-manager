package main

import (
	"context"
	"fmt"
	"fund-manager/config"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	if err := config.LoadEnv(); err != nil {
		log.Fatal("Failed to load env:", err)
	}

	pool, queries, err := config.InitDatabase(ctx)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer pool.Close()

	stocks, err := queries.GetStocks(ctx)
	if err != nil {
		log.Fatalf("Failed to get stocks: %v", err)
	}

	fmt.Printf("Downloading daily OHLC for %d stocks...\n", len(stocks))
	upstox_api := os.Getenv("UPSTOX_API_KEY")
	upstoxUrl := fmt.Sprintf("https://api.upstox.com/v2/login/authorization/dialog?response_type=code&client_id=%s&redirect_uri=%s&state=%s", upstox_api, "http://localhost:3000", "")
	fmt.Println(upstoxUrl)

}
