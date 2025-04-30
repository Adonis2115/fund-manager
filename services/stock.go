package services

import (
	"context"
	"fund-manager/internal/repository"
	initializers "fund-manager/utils"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func GetStockList() []repository.Stock {
	ctx := context.Background()
	stockList, err := initializers.Queries.GetStocks(ctx)

	if err != nil {
		log.Fatal(err)
	}
	return stockList
}

func GetTopStocksByReturn() []repository.GetTopStocksByReturnRow {
	ctx := context.Background()
	givenDate := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	var ts pgtype.Timestamp
	err := ts.Scan(givenDate)
	if err != nil {
		log.Fatalf("Failed to scan timestamp: %v", err)
	}
	listInput := repository.GetTopStocksByReturnParams{Column1: ts, Column2: 12, Column3: "large", Limit: 10}
	stockListByReturn, err := initializers.Queries.GetTopStocksByReturn(ctx, listInput)
	if err != nil {
		log.Fatal(err)
	}
	return stockListByReturn
}
