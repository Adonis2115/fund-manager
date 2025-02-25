package services

import (
	"context"
	"fund-manager/internal/repository"
	initializers "fund-manager/utils"
	"log"
)

func GetStockList() []repository.Stock {
	ctx := context.Background()
	stockList, err := initializers.Queries.GetStocks(ctx)

	if err != nil {
		log.Fatal(err)
	}
	return stockList
}
