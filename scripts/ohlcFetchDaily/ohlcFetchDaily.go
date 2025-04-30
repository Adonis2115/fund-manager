package main

import (
	"context"
	"fmt"
	config "fund-manager/config"
	"fund-manager/internal/repository"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/shopspring/decimal"
)

func init() {
	config.LoadEnv()
}

func main() {
	ctx := context.Background()
	connStr := os.Getenv("POSTGRES")
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer conn.Close(ctx)
	queries := repository.New(conn)

	stocks, err := queries.GetStocks(ctx)
	if err != nil {
		fmt.Println(err)
	}
	now := time.Now()
	for i, stock := range stocks {
		fmt.Printf("%d of %d\n", i, len(stocks)-1)
		ohlcRecords := []repository.BulkCreateDailyParams{}
		params := &chart.Params{
			Symbol:   stock.Customsymbol,
			Start:    &datetime.Datetime{Month: 1, Day: 1, Year: 2021},
			End:      &datetime.Datetime{Month: int(now.Month()), Day: int(now.Day()) - 1, Year: int(now.Year())},
			Interval: datetime.OneDay,
		}
		iter := chart.Get(params)
		for iter.Next() {
			bar := iter.Bar()
			open, err := DecimalToPgNumeric(bar.Open)
			if err != nil {
				fmt.Println(err)
			}
			high, err := DecimalToPgNumeric(bar.High)
			if err != nil {
				fmt.Println(err)
			}
			low, err := DecimalToPgNumeric(bar.Low)
			if err != nil {
				fmt.Println(err)
			}
			close, err := DecimalToPgNumeric(bar.Close)
			if err != nil {
				fmt.Println(err)
			}
			adjClose, err := DecimalToPgNumeric(bar.AdjClose)
			if err != nil {
				fmt.Println(err)
			}
			timestamp, err := IntToPgTimestamp(int64(bar.Timestamp))
			if err != nil {
				fmt.Println(err)
			}
			chartRecord := repository.BulkCreateDailyParams{
				ID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
				Stockid:  stock.ID,
				Open:     open,
				High:     high,
				Low:      low,
				Close:    close,
				Adjclose: adjClose,
				Volume: pgtype.Int4{
					Int32: int32(bar.Volume),
					Valid: true,
				},
				Timestamp: timestamp,
			}
			ohlcRecords = append(ohlcRecords, chartRecord)
		}
		queries.BulkCreateDaily(ctx, ohlcRecords)
		if err := iter.Err(); err != nil {
			fmt.Printf("%s for id %s was unable to fetch record.\n", stock.Symbol, stock.ID)
		}
	}
	fmt.Println("Downloaded dailies data")
}

func DecimalToPgNumeric(d decimal.Decimal) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(d.String())
	return n, err
}

func IntToPgTimestamp(unixTime int64) (pgtype.Timestamp, error) {
	t := time.Unix(unixTime, 0) // convert int (seconds) to time.Time
	var ts pgtype.Timestamp
	err := ts.Scan(t)
	return ts, err
}
