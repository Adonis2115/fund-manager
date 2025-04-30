package services

import (
	"context"
	"fund-manager/internal/repository"
)

type QueryInterface interface {
	GetStocks(ctx context.Context) ([]repository.Stock, error)
	GetTopStocksByReturn(ctx context.Context, input repository.GetTopStocksByReturnParams) ([]repository.GetTopStocksByReturnRow, error)
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
