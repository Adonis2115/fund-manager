package main

import (
	"context"
	"fmt"
	"fund-manager/config"
	"fund-manager/internal/repository"
	"fund-manager/internal/services"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func main() {
	ctx := context.Background()

	if err := config.LoadEnv(); err != nil {
		log.Fatal("Failed to load env:", err)
	}

	pool, queries, err := config.InitDatabase(ctx)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer pool.Close()

	// inject queries into service
	svc := services.NewService(queries)

	givenDate := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	var ts pgtype.Timestamp
	if err := ts.Scan(givenDate); err != nil {
		log.Fatalf("Failed to scan timestamp: %v", err)
	}

	input := repository.GetTopStocksByReturnParams{
		Column1: ts,
		Column2: 12,
		Column3: "all",
		Limit:   10,
	}

	stockList, err := svc.GetTopStocksByReturn(ctx, input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stockList)
}
