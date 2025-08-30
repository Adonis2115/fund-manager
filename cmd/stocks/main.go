package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"fund-manager/config"
	"fund-manager/internal/repository"
	"fund-manager/internal/services"
	"log"
	"os"
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

	givenDate := time.Date(2025, 8, 14, 0, 0, 0, 0, time.UTC)
	var ts pgtype.Timestamp
	if err := ts.Scan(givenDate); err != nil {
		log.Fatalf("Failed to scan timestamp: %v", err)
	}

	input := repository.GetTopStocksByReturnParams{
		Column1: ts,
		Column2: 12,
		Column3: []string{"mid", "small", "micro"},
		Limit:   10,
	}

	stockList, err := svc.GetTopStocksByReturn(ctx, input)
	if err != nil {
		log.Fatal(err)
	}

	err = exportStockListToCSV("stockList.csv", stockList)
	if err != nil {
		log.Fatalf("Failed to export trade logs: %v", err)
	}
	fmt.Println("Trade logs exported to trade_logs.csv")
}

func exportStockListToCSV(filename string, stockList []repository.GetTopStocksByReturnRow) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, stock := range stockList {
		record := []string{
			stock.Symbol,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
