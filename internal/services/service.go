package services

import (
	"context"
	"fund-manager/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type QueryInterface interface {
	GetStocks(ctx context.Context) ([]repository.Stock, error)
	GetTopStocksByReturn(ctx context.Context, input repository.GetTopStocksByReturnParams) ([]repository.GetTopStocksByReturnRow, error)
	GetLatestClosePrice(ctx context.Context, input repository.GetLatestClosePriceParams) (pgtype.Numeric, error) // ✅ Add this line
}

type Service struct {
	Queries QueryInterface
}

func NewService(queries *repository.Queries) *Service {
	return &Service{Queries: queries}
}

func (s *Service) GetTopStocksByReturn(ctx context.Context, input repository.GetTopStocksByReturnParams) ([]repository.GetTopStocksByReturnRow, error) {
	return s.Queries.GetTopStocksByReturn(ctx, input)
}

func (s *Service) GetStockList(ctx context.Context) ([]repository.Stock, error) {
	return s.Queries.GetStocks(ctx)
}

func (s *Service) GetLatestClose(ctx context.Context, input repository.GetLatestClosePriceParams) (pgtype.Numeric, error) {
	return s.Queries.GetLatestClosePrice(ctx, input)
}
