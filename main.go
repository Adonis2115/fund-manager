package main

import (
	"fund-manager/services"
	"log"

	_ "github.com/lib/pq"
)

func init() {
	services.ConnectToDb()
}

func main() {
	// stockls := services.AddStock()
	stockls := services.GetStockList()
	log.Println(stockls)
}
