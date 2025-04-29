package main

import (
	initializers "fund-manager/utils"
)

func init() {
	initializers.ConnectToDb()
}

func main() {
	// stockls := services.AddStock()
	// stockls := services.GetStockList()
	// log.Println(stockls)
}
