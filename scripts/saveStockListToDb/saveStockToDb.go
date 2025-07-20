package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"fund-manager/config"
	"fund-manager/internal/repository"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func main() {
	ctx := context.Background()

	// Load .env and connect to DB
	if err := config.LoadEnv(); err != nil {
		log.Fatal("Failed to load env:", err)
	}

	pool, queries, err := config.InitDatabase(ctx)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer pool.Close()

	// Read all CSV files in /data/stocks
	files, err := os.ReadDir("data/stocks")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !file.Type().IsRegular() || !strings.HasSuffix(file.Name(), ".csv") {
			continue
		}

		processStockCSV(ctx, queries, file.Name())
	}
}

func processStockCSV(ctx context.Context, queries *repository.Queries, filename string) {
	filePath := "data/stocks/" + filename

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file %s: %v", filename, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV %s: %v", filename, err)
	}

	// Read F&O list
	fnoMap := loadFnoMap()

	stockType := strings.Split(filename, ".")[0]
	var stocks []repository.BulkCreateStocksParams

	for _, record := range data {
		if len(record) < 5 {
			continue // skip invalid rows
		}
		isFno := fnoMap[strings.TrimSpace(record[2])]
		if !strings.HasPrefix(record[4], "INE") {
			continue // skip invalid ISINs
		}

		stocks = append(stocks, repository.BulkCreateStocksParams{
			ID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Name:       record[0],
			Symbol:     record[2],
			Scripttype: stockType,
			Industry:   pgtype.Text{String: record[1], Valid: true},
			Isin:       pgtype.Text{String: record[4], Valid: true},
			Fno:        isFno,
		})
	}

	if len(stocks) == 0 {
		fmt.Printf("No valid stocks found in %s\n", filename)
		return
	}

	result, err := queries.BulkCreateStocks(ctx, stocks)
	if err != nil {
		log.Fatalf("Failed to insert stocks from %s: %v", filename, err)
	}

	fmt.Printf("Inserted %d stocks from %s\n", result, filename)
}

func loadFnoMap() map[string]bool {
	file, err := os.Open("data/fnoList.csv")
	if err != nil {
		log.Fatalf("Failed to open FNO list: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	fnoList, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read FNO CSV: %v", err)
	}

	fnoMap := make(map[string]bool)
	for _, fno := range fnoList {
		if len(fno) >= 2 {
			fnoMap[strings.TrimSpace(fno[1])] = true
		}
	}
	return fnoMap
}
