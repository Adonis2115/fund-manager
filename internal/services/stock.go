package services

import (
	"context"
	config "fund-manager/config"
	"fund-manager/internal/repository"
	"log"
)

func GetStockList() []repository.Stock {
	ctx := context.Background()
	stockList, err := config.Queries.GetStocks(ctx)

	if err != nil {
		log.Fatal(err)
	}
	return stockList
}

func GetTopStocksByReturn(inputTop repository.GetTopStocksByReturnParams) []repository.GetTopStocksByReturnRow {
	ctx := context.Background()
	stockListByReturn, err := config.Queries.GetTopStocksByReturn(ctx, inputTop)
	if err != nil {
		log.Fatal(err)
	}
	return stockListByReturn
}
