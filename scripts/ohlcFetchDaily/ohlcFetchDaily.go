package main

import (
	"context"
	"fmt"
	"fund-manager/config"
	"fund-manager/internal/repository"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/shopspring/decimal"
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
		if err := fetchAndStoreDailyData(ctx, queries, stock); err != nil {
			log.Printf("Error processing %s: %v", stock.Symbol, err)
		}
	}

	fmt.Println("✅ All daily OHLC data downloaded and stored.")
}

func fetchAndStoreDailyData(ctx context.Context, queries *repository.Queries, stock repository.Stock) error {
	now := time.Now()
	params := &chart.Params{
		Symbol:   stock.Customsymbol,
		Start:    &datetime.Datetime{Month: 1, Day: 1, Year: 2021},
		End:      &datetime.Datetime{Month: int(now.Month()), Day: int(now.Day() - 1), Year: int(now.Year())},
		Interval: datetime.OneDay,
	}

	iter := chart.Get(params)
	var ohlcRecords []repository.BulkCreateDailyParams

	for iter.Next() {
		bar := iter.Bar()

		open, err := DecimalToPgNumeric(bar.Open)
		if err != nil {
			log.Printf("Invalid Open for %s: %v", stock.Symbol, err)
			continue
		}
		high, err := DecimalToPgNumeric(bar.High)
		if err != nil {
			log.Printf("Invalid High for %s: %v", stock.Symbol, err)
			continue
		}
		low, err := DecimalToPgNumeric(bar.Low)
		if err != nil {
			log.Printf("Invalid Low for %s: %v", stock.Symbol, err)
			continue
		}
		closePrice, err := DecimalToPgNumeric(bar.Close)
		if err != nil {
			log.Printf("Invalid Close for %s: %v", stock.Symbol, err)
			continue
		}
		adjClose, err := DecimalToPgNumeric(bar.AdjClose)
		if err != nil {
			log.Printf("Invalid AdjClose for %s: %v", stock.Symbol, err)
			continue
		}
		timestamp, err := IntToPgTimestamp(int64(bar.Timestamp))
		if err != nil {
			log.Printf("Invalid Timestamp for %s: %v", stock.Symbol, err)
			continue
		}

		record := repository.BulkCreateDailyParams{
			ID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Stockid:  stock.ID,
			Open:     open,
			High:     high,
			Low:      low,
			Close:    closePrice,
			Adjclose: adjClose,
			Volume: pgtype.Int4{
				Int32: int32(bar.Volume),
				Valid: true,
			},
			Timestamp: timestamp,
		}
		ohlcRecords = append(ohlcRecords, record)
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to fetch chart data: %w", err)
	}

	if len(ohlcRecords) == 0 {
		log.Printf("No OHLC data found for %s", stock.Symbol)
		return nil
	}

	_, err := queries.BulkCreateDaily(ctx, ohlcRecords)
	if err != nil {
		return fmt.Errorf("failed to insert OHLC for %s: %w", stock.Symbol, err)
	}

	return nil
}

func DecimalToPgNumeric(d decimal.Decimal) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(d.String())
	return n, err
}

func IntToPgTimestamp(unixTime int64) (pgtype.Timestamp, error) {
	t := time.Unix(unixTime, 0)
	var ts pgtype.Timestamp
	err := ts.Scan(t)
	return ts, err
}
