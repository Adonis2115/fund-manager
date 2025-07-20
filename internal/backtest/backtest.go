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
	Symbol      string
	EntryDate   time.Time
	ExitDate    time.Time
	EntryPrice  float64
	ExitPrice   float64
	Profit      float64
	ProfitPct   float64
	DaysHeld    int
	Quantity    float64
	AmountUsed  float64
	MaxDrawdown float64 // New field for individual stock drawdown
}

type BacktestResult struct {
	TradeLogs      []TradeLog
	EquityCurve    []float64
	MonthlyReturns []float64
	Drawdown       float64
	CAGR           float64
	PortfolioLog   [][]string
	TotalTrades    int
	WinningTrades  int
	WinRate        float64
	AverageProfit  float64
	Profit         float64
}

func RunBacktest(ctx context.Context, cfg BacktestConfig) BacktestResult {
	equity := cfg.InitialCapital
	equityCurve := make([]float64, 0)
	monthlyReturns := make([]float64, 0)
	portfolioLog := make([][]string, 0)

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
		monthProfit := 0.0

		for _, row := range rows {
			price := getLatestClose(ctx, cfg.Service, row.Symbol, monthDate)
			if _, held := currentPortfolio[row.Symbol]; !held {
				entryPrices[row.Symbol] = price
				entryDates[row.Symbol] = monthDate
			}
			newPortfolio[row.Symbol] = struct{}{}
			currentSymbols = append(currentSymbols, row.Symbol)
		}

		// Exit stocks not in newPortfolio
		for sym := range currentPortfolio {
			if _, stillHeld := newPortfolio[sym]; !stillHeld {
				exitPrice := getLatestClose(ctx, cfg.Service, sym, monthDate)
				entryPrice := entryPrices[sym]
				daysHeld := int(monthDate.Sub(entryDates[sym]).Hours() / 24)
				alloc := cfg.InitialCapital / float64(cfg.TopN)
				quantity := math.Floor(alloc / entryPrice)
				amount := quantity * entryPrice
				profit := (exitPrice - entryPrice) * quantity
				profitPct := (profit / amount) * 100

				tradeLogs = append(tradeLogs, TradeLog{
					Symbol:      sym,
					EntryDate:   entryDates[sym],
					ExitDate:    monthDate,
					EntryPrice:  entryPrice,
					ExitPrice:   exitPrice,
					Profit:      profit,
					ProfitPct:   profitPct,
					DaysHeld:    daysHeld,
					Quantity:    quantity,
					AmountUsed:  amount,
					MaxDrawdown: getStockDrawdown(ctx, cfg, sym, entryDates[sym], monthDate),
				})

				monthProfit += profit
				delete(entryPrices, sym)
				delete(entryDates, sym)
			}
		}

		equity += monthProfit
		equityCurve = append(equityCurve, equity)
		monthlyReturns = append(monthlyReturns, monthProfit/cfg.InitialCapital)
		portfolioLog = append(portfolioLog, currentSymbols)
		currentPortfolio = newPortfolio
	}

	// Final exits
	for sym := range currentPortfolio {
		entryPrice := entryPrices[sym]
		exitPrice := getLatestClose(ctx, cfg.Service, sym, cfg.EndDate)
		daysHeld := int(cfg.EndDate.Sub(entryDates[sym]).Hours() / 24)
		alloc := cfg.InitialCapital / float64(cfg.TopN)
		quantity := math.Floor(alloc / entryPrice)
		amount := quantity * entryPrice
		profit := (exitPrice - entryPrice) * quantity
		profitPct := (profit / amount) * 100

		tradeLogs = append(tradeLogs, TradeLog{
			Symbol:      sym,
			EntryDate:   entryDates[sym],
			ExitDate:    cfg.EndDate,
			EntryPrice:  entryPrice,
			ExitPrice:   exitPrice,
			Profit:      profit,
			ProfitPct:   profitPct,
			DaysHeld:    daysHeld,
			Quantity:    quantity,
			AmountUsed:  amount,
			MaxDrawdown: getStockDrawdown(ctx, cfg, sym, entryDates[sym], cfg.EndDate),
		})

		equity += profit
		equityCurve = append(equityCurve, equity)
	}

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

	months := int(cfg.EndDate.Sub(cfg.StartDate).Hours() / (24 * 30))
	cagr := computeCAGR(cfg.InitialCapital, equity, months)
	drawdown := maxDrawdown(equityCurve)

	return BacktestResult{
		TradeLogs:      tradeLogs,
		EquityCurve:    equityCurve,
		MonthlyReturns: monthlyReturns,
		Drawdown:       drawdown,
		CAGR:           cagr,
		PortfolioLog:   portfolioLog,
		TotalTrades:    total,
		WinningTrades:  wins,
		WinRate:        winRate,
		AverageProfit:  avgProfit,
		Profit:         sumProfits,
	}
}

func toPgDate(t time.Time) (d pgtype.Date) {
	dateOnly := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	_ = d.Scan(dateOnly)
	return
}

func toPgTimestamp(t time.Time) (ts pgtype.Timestamp) {
	_ = ts.Scan(t)
	return
}

func getLatestClose(ctx context.Context, s *services.Service, symbol string, date time.Time) float64 {
	input := repository.GetLatestClosePriceParams{
		Symbol:    symbol,
		Timestamp: toPgDate(date),
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

func getStockDrawdown(ctx context.Context, cfg BacktestConfig, symbol string, start, end time.Time) float64 {
	query := repository.GetHistoricalStockPricesParams{
		Symbol:      symbol,
		Timestamp:   toPgDate(start),
		Timestamp_2: toPgDate(end),
	}
	prices, err := cfg.Service.GetStockPrices(ctx, query)
	if err != nil || len(prices) == 0 {
		log.Printf("Error fetching prices for drawdown: %s: %v", symbol, err)
		return 0
	}

	peakF, err := prices[0].Close.Float64Value()
	if err != nil {
		log.Printf("Invalid peak value for %s: %v", symbol, err)
		return 0
	}
	peak := peakF.Float64
	maxDD := 0.0
	for _, p := range prices {
		closeF, err := p.Close.Float64Value()
		if err != nil {
			log.Printf("Invalid price for %s: %v", symbol, err)
			continue
		}
		price := closeF.Float64
		if price > peak {
			peak = price
		}
		drawdown := (peak - price) / peak
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
