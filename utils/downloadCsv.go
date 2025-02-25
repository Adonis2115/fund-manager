package initializers

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type StockList struct {
	cap string
	url string
}

func DownloadStocksCsv() {
	stockList := []StockList{{cap: "large", url: "https://nsearchives.nseindia.com/content/indices/ind_nifty100list.csv"},
		{cap: "mid", url: "https://nsearchives.nseindia.com/content/indices/ind_niftymidcap150list.csv"},
		{cap: "small", url: "https://nsearchives.nseindia.com/content/indices/ind_niftysmallcap250list.csv"},
		{cap: "micro", url: "https://nsearchives.nseindia.com/content/indices/ind_niftymicrocap250_list.csv"}}
	for _, segment := range stockList {
		fmt.Printf("Fetching %s\n", segment.cap)
		fileName := "./data/stocks/" + segment.cap + ".csv"
		err := DownloadCSV(segment.url, fileName)
		if err != nil {
			fmt.Printf("Error fetching %s\n", segment.cap)
			fmt.Println(err)
		}
	}
	fmt.Println("FNO List")
	fileName := "./data/" + "fnoList" + ".csv"
	err := DownloadCSV("https://nsearchives.nseindia.com/content/fo/fo_mktlots.csv", fileName)
	if err != nil {
		fmt.Println("Error fetching FNO List")
		fmt.Println(err)
	}
	fmt.Println("Done fetching all csv files.")
}

func DownloadCSV(link, filename string) error {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
