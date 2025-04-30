package main

import (
	"context"
	"encoding/csv"
	"fmt"
	config "fund-manager/config"
	"fund-manager/internal/repository"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

	folder := "data/stocks"
	files, err := os.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.Type().IsRegular() && file.Name()[len(file.Name())-4:] == ".csv" {
			f, err := os.Open(folder + "/" + file.Name())
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			csvReader := csv.NewReader(f)
			data, err := csvReader.ReadAll()
			if err != nil {
				log.Fatal(err)
			}
			stockType := strings.Split(file.Name(), ".")[0]
			fnoFile, err := os.Open("data/fnoList.csv")
			if err != nil {
				log.Fatal(err)
			}
			defer fnoFile.Close()
			fnoCsvReader := csv.NewReader(fnoFile)
			fnoList, err := fnoCsvReader.ReadAll()
			if err != nil {
				log.Fatal(err)
			}
			var stocks []repository.BulkCreateStocksParams
			for _, record := range data {
				isFno := false
				for _, fno := range fnoList {
					if strings.TrimSpace(fno[1]) == record[2] {
						isFno = true
						break
					}
				}
				if strings.HasPrefix(record[4], "INE") {
					stock := repository.BulkCreateStocksParams{
						ID:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
						Name:         record[0],
						Symbol:       record[2],
						Customsymbol: record[2] + ".NS",
						Scripttype:   stockType,
						Industry:     pgtype.Text{String: record[1], Valid: true},
						Isin:         pgtype.Text{String: record[4], Valid: true},
						Fno:          isFno,
					}
					stocks = append(stocks, stock)
				}
			}
			result, err := queries.BulkCreateStocks(ctx, stocks)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(result)
		}
	}
}
