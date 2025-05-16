package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Instrument struct {
	Segment        string  `json:"segment"`
	Name           string  `json:"name"`
	Exchange       string  `json:"exchange"`
	ISIN           string  `json:"isin,omitempty"`
	InstrumentType string  `json:"instrument_type"`
	InstrumentKey  string  `json:"instrument_key"`
	LotSize        int     `json:"lot_size"`
	FreezeQuantity float64 `json:"freeze_quantity"`
	ExchangeToken  string  `json:"exchange_token"`
	TickSize       float64 `json:"tick_size"`
	TradingSymbol  string  `json:"trading_symbol"`
	QtyMultiplier  float64 `json:"qty_multiplier"`
	SecurityType   string  `json:"security_type,omitempty"`
}

func main() {
	// Load the JSON file
	fileBytes, err := os.ReadFile("data/NSE.json") // replace with your filename
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Unmarshal into a slice of instruments
	var instruments []Instrument
	if err := json.Unmarshal(fileBytes, &instruments); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Filter NSE_EQ instruments
	var nseEq []Instrument
	for _, inst := range instruments {
		if inst.Segment == "NSE_EQ" {
			nseEq = append(nseEq, inst)
		}
	}

	// Marshal the filtered result
	output, err := json.MarshalIndent(nseEq, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Write to file
	if err := os.WriteFile("data/nse_eq_only.json", output, 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	fmt.Printf("✅ Filtered %d NSE_EQ instruments and saved to nse_eq_only.json\n", len(nseEq))
}
