package main

import (
	"context"
	"fmt"
	"fund-manager/config"
	"fund-manager/internal/backtest"
	"fund-manager/internal/services"
	"log"
	"time"
)

func main() {
	ctx := context.Background()

	// Load env and connect DB
	if err := config.LoadEnv(); err != nil {
		log.Fatal("Failed to load env:", err)
	}
	pool, queries, err := config.InitDatabase(ctx)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer pool.Close()

	// Build service layer
	service := services.NewService(queries)

	// Configure backtest
	cfg := backtest.BacktestConfig{
		StartDate:      time.Date(2021, 9, 23, 0, 0, 0, 0, time.UTC),
		EndDate:        time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
		Months:         12,
		TopN:           10,
		ScriptType:     "large",
		InitialCapital: 1000000,
		Service:        service,
	}

	// Run it
	result := backtest.RunBacktest(ctx, cfg)

	// Display result
	fmt.Printf("Final Equity: %.2f\n", result.EquityCurve[len(result.EquityCurve)-1])
	fmt.Printf("Max Drawdown: %.2f%%\n", result.Drawdown*100)
	fmt.Printf("CAGR: %.2f%%\n", result.CAGR*100)
}
