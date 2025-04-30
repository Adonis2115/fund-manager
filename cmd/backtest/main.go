package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"fund-manager/config"
	"fund-manager/internal/backtest"
	"fund-manager/internal/services"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	ctx := context.Background()
	config.LoadEnv()
	pool, queries, err := config.InitDatabase(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer pool.Close()

	service := services.NewService(queries)

	cfg := backtest.BacktestConfig{
		StartDate:      time.Date(2025, 2, 19, 0, 0, 0, 0, time.UTC),
		EndDate:        time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
		TopN:           10,
		ScriptType:     "all",
		InitialCapital: 1000000,
		Service:        service,
	}

	result := backtest.RunBacktest(ctx, cfg)
	fmt.Printf("Backtest completed.\n")
	fmt.Printf("CAGR: %.2f%%\n", result.CAGR*100)
	fmt.Printf("Max Drawdown: %.2f%%\n", result.Drawdown*100)
	fmt.Printf("Total Trades: %d\n", result.TotalTrades)
	fmt.Printf("Winning Trades: %d\n", result.WinningTrades)
	fmt.Printf("Win Rate: %.2f%%\n", result.WinRate*100)
	fmt.Printf("Average Profit: %.2f\n", result.AverageProfit)
	fmt.Printf("Net Profit: %.2f\n", result.EquityCurve[len(result.EquityCurve)-1]-cfg.InitialCapital)

	err = exportTradeLogsToCSV("trade_logs.csv", result.TradeLogs)
	if err != nil {
		log.Fatalf("Failed to export trade logs: %v", err)
	}
	fmt.Println("Trade logs exported to trade_logs.csv")
}

func exportTradeLogsToCSV(filename string, trades []backtest.TradeLog) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Symbol", "EntryDate", "ExitDate", "EntryPrice", "ExitPrice", "Profit", "ProfitPct", "DaysHeld", "Quantity", "AmountUsed"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, trade := range trades {
		record := []string{
			trade.Symbol,
			trade.EntryDate.Format("2006-01-02"),
			trade.ExitDate.Format("2006-01-02"),
			fmt.Sprintf("%.2f", trade.EntryPrice),
			fmt.Sprintf("%.2f", trade.ExitPrice),
			fmt.Sprintf("%.2f", trade.Profit),
			fmt.Sprintf("%.2f", trade.ProfitPct),
			strconv.Itoa(trade.DaysHeld),
			fmt.Sprintf("%.0f", trade.Quantity),
			fmt.Sprintf("%.2f", trade.AmountUsed),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
