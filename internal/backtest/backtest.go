// 📁 internal/backtest/backtest.go
package backtest

import (
	"context"
	"fund-manager/internal/repository"
	"fund-manager/internal/services"
	"log"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/exp/slices"
)

type BacktestConfig struct {
	StartDate      time.Time
	EndDate        time.Time
	TopN           int32
	ScriptType     string
	InitialCapital float64
	Service        *services.Service
}

type TradeLog struct {
	Symbol     string
	EntryDate  time.Time
	ExitDate   time.Time
	EntryPrice float64
	ExitPrice  float64
	Profit     float64
	ProfitPct  float64
	DaysHeld   int
	Quantity   float64
	AmountUsed float64
}

type BacktestResult struct {
	TotalTrades    int
	WinningTrades  int
	WinRate        float64
	AverageProfit  float64
	TradeLogs      []TradeLog
	EquityCurve    []float64
	MonthlyReturns []float64
	Drawdown       float64
	CAGR           float64
	PortfolioLog   [][]string
}

func RunBacktest(ctx context.Context, cfg BacktestConfig) BacktestResult {
	equity := cfg.InitialCapital
	equityCurve := make([]float64, 0)
	monthlyReturns := make([]float64, 0)
	portfolioLog := make([][]string, 0)

	// Removed: var previousPrices = make(map[string]float64)
	tradeLogs := make([]TradeLog, 0)

	currentPortfolio := make(map[string]struct{})
	entryPrices := make(map[string]float64)
	entryDates := make(map[string]time.Time)

	for month := 0; ; month++ {
		monthDate := cfg.StartDate.AddDate(0, month, 0)
		if monthDate.After(cfg.EndDate) {
			break
		}
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

		newPortfolio := make(map[string]struct{})
		currentSymbols := make([]string, 0, len(rows))

		monthlyReturn := 0.0
		count := 0
		for _, row := range rows {
			newPortfolio[row.Symbol] = struct{}{}
			currentSymbols = append(currentSymbols, row.Symbol)

			price := getLatestClose(ctx, cfg.Service, row.Symbol, monthDate)
			if _, held := currentPortfolio[row.Symbol]; !held {
				entryPrices[row.Symbol] = price
				entryDates[row.Symbol] = monthDate
			}

			newPortfolio[row.Symbol] = struct{}{}
			currentSymbols = append(currentSymbols, row.Symbol)
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
		// Exit stocks not in newPortfolio
		for sym := range currentPortfolio {
			if _, stillHeld := newPortfolio[sym]; !stillHeld {
				exitPrice := getLatestClose(ctx, cfg.Service, sym, monthDate)
				entryPrice := entryPrices[sym]
				daysHeld := int(monthDate.Sub(entryDates[sym]).Hours() / 24)
				profit := exitPrice - entryPrice
				profitPct := (profit / entryPrice) * 100
				alloc := cfg.InitialCapital / float64(cfg.TopN)
				quantity := math.Floor(alloc / entryPrice)
				amount := quantity * entryPrice
				tradeLogs = append(tradeLogs, TradeLog{
					Symbol:     sym,
					EntryDate:  entryDates[sym],
					ExitDate:   monthDate,
					EntryPrice: entryPrice,
					ExitPrice:  exitPrice,
					Profit:     profit,
					ProfitPct:  profitPct,
					DaysHeld:   daysHeld,
					Quantity:   quantity,
					AmountUsed: amount,
				})
				delete(entryPrices, sym)
				delete(entryDates, sym)
			}
		}
		currentPortfolio = newPortfolio
	}

	// Log open positions at end of backtest
	for sym := range currentPortfolio {
		entryPrice := entryPrices[sym]
		exitPrice := getLatestClose(ctx, cfg.Service, sym, cfg.EndDate)
		daysHeld := int(cfg.EndDate.Sub(entryDates[sym]).Hours() / 24)
		profit := exitPrice - entryPrice
		profitPct := (profit / entryPrice) * 100
		alloc := cfg.InitialCapital / float64(cfg.TopN)
		quantity := math.Floor(alloc / entryPrice)
		amount := quantity * entryPrice

		tradeLogs = append(tradeLogs, TradeLog{
			Symbol:     sym,
			EntryDate:  entryDates[sym],
			ExitDate:   cfg.EndDate,
			EntryPrice: entryPrice,
			ExitPrice:  exitPrice,
			Profit:     profit,
			ProfitPct:  profitPct,
			DaysHeld:   daysHeld,
			Quantity:   quantity,
			AmountUsed: amount,
		})
	}

	dd := maxDrawdown(equityCurve)
	months := int(cfg.EndDate.Sub(cfg.StartDate).Hours() / (24 * 30))
	cagr := computeCAGR(cfg.InitialCapital, equity, months)

	// Sort trade logs by EntryDate
	slices.SortFunc(tradeLogs, func(a, b TradeLog) int {
		if a.EntryDate.Before(b.EntryDate) {
			return -1
		} else if a.EntryDate.After(b.EntryDate) {
			return 1
		}
		return 0
	})

	// Calculate stats
	total := len(tradeLogs)
	wins := 0
	sumProfits := 0.0
	for _, t := range tradeLogs {
		if t.Profit > 0 {
			wins++
		}
		sumProfits += t.Profit
	}
	winRate := 0.0
	avgProfit := 0.0
	if total > 0 {
		winRate = float64(wins) / float64(total)
		avgProfit = sumProfits / float64(total)
	}

	return BacktestResult{
		TradeLogs:      tradeLogs,
		EquityCurve:    equityCurve,
		MonthlyReturns: monthlyReturns,
		Drawdown:       dd,
		CAGR:           cagr,
		PortfolioLog:   portfolioLog,
		TotalTrades:    total,
		WinningTrades:  wins,
		WinRate:        winRate,
		AverageProfit:  avgProfit,
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
