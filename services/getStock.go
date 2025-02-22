package services

import (
	"context"
	"fund-manager/internal/repository"
	"log"

	_ "github.com/lib/pq"
)

func GetStock() []repository.Stock {
	ctx := context.Background()
	stockList, err := Queries.GetStocks(ctx)

	if err != nil {
		log.Fatal(err)
	}
	return stockList
}
