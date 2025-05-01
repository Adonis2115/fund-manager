package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type StockSegment struct {
	Cap string
	URL string
}

var stockList = []StockSegment{
	{Cap: "large", URL: "https://nsearchives.nseindia.com/content/indices/ind_nifty100list.csv"},
	{Cap: "mid", URL: "https://nsearchives.nseindia.com/content/indices/ind_niftymidcap150list.csv"},
	{Cap: "small", URL: "https://nsearchives.nseindia.com/content/indices/ind_niftysmallcap250list.csv"},
	// {Cap: "micro", URL: "https://nsearchives.nseindia.com/content/indices/ind_niftymicrocap250_list.csv"},
}

const dataDir = "data"
const stocksDir = "data/stocks"

func main() {
	ensureDir(stocksDir)

	for _, segment := range stockList {
		filePath := filepath.Join(stocksDir, segment.Cap+".csv")
		log.Printf("📥 Fetching %s cap stock list...", segment.Cap)
		if err := downloadCSV(segment.URL, filePath); err != nil {
			log.Printf("❌ Failed to download %s cap list: %v", segment.Cap, err)
			continue
		}
		log.Printf("✅ Downloaded %s → %s", segment.Cap, filePath)
	}

	log.Println("📥 Fetching FNO list...")
	fnoPath := filepath.Join(dataDir, "fnoList.csv")
	if err := downloadCSV("https://nsearchives.nseindia.com/content/fo/fo_mktlots.csv", fnoPath); err != nil {
		log.Fatalf("❌ Failed to download FNO list: %v", err)
	}
	log.Printf("✅ Downloaded FNO list → %s", fnoPath)

	log.Println("🎉 Done fetching all CSV files.")
}

func downloadCSV(url, filename string) error {
	// Force HTTP/1.1 by using a custom transport
	client := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: false,
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("request creation failed: %w", err)
	}

	// NSE requires these headers to respond properly
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0")
	req.Header.Set("Referer", "https://www.nseindia.com/")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received HTTP status %s", resp.Status)
	}

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("file creation failed: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("file write failed: %w", err)
	}

	return nil
}

func ensureDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatalf("❌ Failed to create directory %s: %v", path, err)
		}
	}
}
