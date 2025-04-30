package backtest

import (
	"context"
	"fund-manager/internal/repository"
	"fund-manager/internal/services"
	"log"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type BacktestConfig struct {
	StartDate      time.Time
	Months         int
	TopN           int32
	ScriptType     string
	InitialCapital float64
	Service        *services.Service
}

type BacktestResult struct {
	EquityCurve    []float64
	MonthlyReturns []float64
	Drawdown       float64
	CAGR           float64
	PortfolioLog   [][]string
}

func RunBacktest(ctx context.Context, cfg BacktestConfig) BacktestResult {
	equity := cfg.InitialCapital
	equityCurve := make([]float64, 0, cfg.Months)
	monthlyReturns := make([]float64, 0, cfg.Months)
	portfolioLog := make([][]string, 0, cfg.Months)

	var previousPrices map[string]float64 = make(map[string]float64)

	for month := 0; month < cfg.Months; month++ {
		monthDate := cfg.StartDate.AddDate(0, month, 0)
		params := repository.GetTopStocksByReturnParams{
			Column1: toPgTimestamp(monthDate),
			Column2: 12,
			Column3: cfg.ScriptType,
			Limit:   cfg.TopN,
		}

		rows, err := cfg.Service.GetTopStocksByReturn(ctx, params)
		if err != nil {
			log.Printf("Error fetching top stocks for %s: %v", monthDate.Format("2006-01-02"), err)
			continue
		}

		newPortfolio := make(map[string]bool)
		currentSymbols := make([]string, 0, len(rows))

		monthlyReturn := 0.0
		count := 0
		for _, row := range rows {
			price := getLatestClose(ctx, cfg.Service, row.Symbol, monthDate)
			prevPrice := previousPrices[row.Symbol]

			if prevPrice > 0 {
				r := (price - prevPrice) / prevPrice
				monthlyReturn += r
				count++
			}

			newPortfolio[row.Symbol] = true
			currentSymbols = append(currentSymbols, row.Symbol)
			previousPrices[row.Symbol] = price
		}

		if count > 0 {
			monthlyReturn = monthlyReturn / float64(count)
		} else {
			monthlyReturn = 0
		}

		equity = equity * (1 + monthlyReturn)
		equityCurve = append(equityCurve, equity)
		monthlyReturns = append(monthlyReturns, monthlyReturn)
		portfolioLog = append(portfolioLog, currentSymbols)
	}

	dd := maxDrawdown(equityCurve)
	cagr := computeCAGR(cfg.InitialCapital, equity, cfg.Months)

	return BacktestResult{
		EquityCurve:    equityCurve,
		MonthlyReturns: monthlyReturns,
		Drawdown:       dd,
		CAGR:           cagr,
		PortfolioLog:   portfolioLog,
	}
}

func toPgTimestamp(t time.Time) (ts pgtype.Timestamp) {
	_ = ts.Scan(t)
	return
}

func getLatestClose(ctx context.Context, s *services.Service, symbol string, date time.Time) float64 {
	input := repository.GetLatestClosePriceParams{
		Symbol:    symbol,
		Timestamp: toPgTimestamp(date),
	}
	result, err := s.GetLatestClose(ctx, input)
	if err != nil || !result.Valid {
		log.Printf("Failed to get latest close for %s at %s: %v", symbol, date.Format("2006-01-02"), err)
		return 0
	}
	f64, err := result.Float64Value()
	if err != nil {
		log.Printf("Error converting pgtype.Numeric to float64 for %s: %v", symbol, err)
		return 0
	}
	return f64.Float64
}

func maxDrawdown(equity []float64) float64 {
	peak := equity[0]
	maxDD := 0.0
	for _, v := range equity {
		if v > peak {
			peak = v
		}
		drawdown := (peak - v) / peak
		if drawdown > maxDD {
			maxDD = drawdown
		}
	}
	return maxDD
}

func computeCAGR(initial, final float64, months int) float64 {
	years := float64(months) / 12.0
	if initial <= 0 || years <= 0 {
		return 0
	}
	return math.Pow(final/initial, 1/years) - 1
}
