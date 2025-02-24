package services

import (
	"context"
	"fund-manager/internal/repository"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func AddStock() repository.Stock {
	ctx := context.Background()
	stock, err := Queries.CreateStock(ctx, repository.CreateStockParams{
		ID:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Name:         "Alpha",
		Symbol:       "ALPH",
		Customsymbol: "NSE:ALPH",
		Scripttype:   "LARGE",
		Industry:     pgtype.Text{String: "test", Valid: true},
		Isin:         pgtype.Text{String: "13214334", Valid: true},
		Fno:          false,
	})

	if err != nil {
		log.Fatal(err)
	}
	return stock
}

func GetStockList() []repository.Stock {
	ctx := context.Background()
	stockList, err := Queries.GetStocks(ctx)

	if err != nil {
		log.Fatal(err)
	}
	return stockList
}
