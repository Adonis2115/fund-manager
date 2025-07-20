package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"fund-manager/config"
	"fund-manager/internal/repository"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	for i, stock := range stocks {
		fmt.Printf("[%d/%d] %s\n", i+1, len(stocks), stock.Symbol)
		if err := importDailyCSVData(ctx, queries, stock); err != nil {
			log.Printf("Error processing %s: %v", stock.Symbol, err)
		}
	}

	fmt.Println("✅ All daily OHLC data downloaded and stored.")
}

func importDailyCSVData(ctx context.Context, queries *repository.Queries, stock repository.Stock) error {
	filePath := filepath.Join("./data/nseDaily/daily", strings.ToLower(stock.Symbol)+".csv")
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	var ohlcRecords []repository.BulkCreateDailyParams
	for idx, row := range records {
		// Skip header if present
		if idx == 0 && strings.ToLower(row[0]) == "date" {
			continue
		}
		if len(row) < 6 {
			log.Printf("Skipping malformed row in %s: %v", stock.Symbol, row)
			continue
		}

		date, err := time.Parse("2006-01-02", row[0])
		if err != nil {
			log.Printf("Invalid date for %s: %v", stock.Symbol, err)
			continue
		}

		open, err := parseToPgNumeric(row[1])
		if err != nil {
			log.Printf("Invalid open for %s: %v", stock.Symbol, err)
			continue
		}
		high, err := parseToPgNumeric(row[2])
		if err != nil {
			log.Printf("Invalid high for %s: %v", stock.Symbol, err)
			continue
		}
		low, err := parseToPgNumeric(row[3])
		if err != nil {
			log.Printf("Invalid low for %s: %v", stock.Symbol, err)
			continue
		}
		closePrice, err := parseToPgNumeric(row[4])
		if err != nil {
			log.Printf("Invalid close for %s: %v", stock.Symbol, err)
			continue
		}
		volInt, err := strconv.Atoi(row[5])
		if err != nil {
			log.Printf("Invalid volume for %s: %v", stock.Symbol, err)
			continue
		}

		record := repository.BulkCreateDailyParams{
			ID:      pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Stockid: stock.ID,
			Open:    open,
			High:    high,
			Low:     low,
			Close:   closePrice,
			Volume: pgtype.Int4{
				Int32: int32(volInt),
				Valid: true,
			},
			Timestamp: pgtype.Date{
				Time:  date,
				Valid: true,
			},
		}
		ohlcRecords = append(ohlcRecords, record)
	}

	if len(ohlcRecords) == 0 {
		log.Printf("No valid rows for %s", stock.Symbol)
		return nil
	}

	_, err = queries.BulkCreateDaily(ctx, ohlcRecords)
	if err != nil {
		return fmt.Errorf("failed to insert OHLC: %w", err)
	}
	return nil
}

func parseToPgNumeric(value string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(value)
	return n, err
}
