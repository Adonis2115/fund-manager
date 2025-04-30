package main

import (
	"fmt"
	config "fund-manager/config"
	"fund-manager/internal/repository"
	"fund-manager/internal/services"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func init() {
	config.ConnectToDb()
}

func main() {
	givenDate := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	var ts pgtype.Timestamp
	err := ts.Scan(givenDate)
	if err != nil {
		log.Fatalf("Failed to scan timestamp: %v", err)
	}
	inputTopStocks := repository.GetTopStocksByReturnParams{Column1: ts, Column2: 12, Column3: "large", Limit: 10}
	stockList := services.GetTopStocksByReturn(inputTopStocks)
	fmt.Println(stockList)
}
