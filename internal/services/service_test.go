package services

import (
	"context"
	"fund-manager/internal/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

type mockQueries struct{}

func (m *mockQueries) GetTopStocksByReturn(ctx context.Context, input repository.GetTopStocksByReturnParams) ([]repository.GetTopStocksByReturnRow, error) {
	return []repository.GetTopStocksByReturnRow{
		{
			ID:               pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Name:             "Infosys",
			Symbol:           "INFY",
			ReturnPercentage: 15,
		},
	}, nil
}

func (m *mockQueries) GetStocks(ctx context.Context) ([]repository.Stock, error) {
	return []repository.Stock{
		{
			ID:     pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Name:   "Reliance",
			Symbol: "RELIANCE",
		},
	}, nil
}

func TestGetTopStocksByReturn(t *testing.T) {
	service := &Service{Queries: &mockQueries{}}
	ctx := context.Background()
	input := repository.GetTopStocksByReturnParams{
		Column1: pgtype.Timestamp{Valid: true},
		Column2: 12,
		Column3: "large",
		Limit:   10,
	}
	result, err := service.GetTopStocksByReturn(ctx, input)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "INFY", result[0].Symbol)
}

func TestGetStockList(t *testing.T) {
	service := &Service{Queries: &mockQueries{}}
	ctx := context.Background()
	result, err := service.GetStockList(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Reliance", result[0].Name)
}
