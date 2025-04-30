package main

import (
	"fmt"
	"fund-manager/services"
	initializers "fund-manager/utils"
)

func init() {
	initializers.ConnectToDb()
}

func main() {
	stockList := services.GetTopStocksByReturn()
	fmt.Println(stockList)
}
