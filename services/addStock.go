package services

import (
	"context"
	"fund-manager/internal/repository"
	"log"

	_ "github.com/lib/pq"
)

func AddStock() repository.Stock {
	ctx := context.Background()
	stock, err := Queries.CreateStock(ctx, repository.CreateStockParams{
		ID:           2,
		Name:         "Alpha",
		Symbol:       "ALPH",
		Customsymbol: "NSE:ALPH",
		Scripttype:   "LARGE",
		Industry:     "test",
		Isin:         "13214334",
		Fno:          false,
	})

	if err != nil {
		log.Fatal(err)
	}
	return stock
}
