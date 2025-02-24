package main

import (
	"fund-manager/services"

	_ "github.com/lib/pq"
)

func init() {
	services.ConnectToDb()
}

func main() {
	// stockls := services.AddStock()
	services.CsvStocks()
	// stockls := services.GetStockList()
	// log.Println(stockls)
}
