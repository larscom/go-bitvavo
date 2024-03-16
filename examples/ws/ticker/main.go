package main

import (
	"log"

	"github.com/larscom/go-bitvavo/v2"
)

func main() {
	markets, err := bitvavo.NewHttpClient().GetMarkets()
	if err != nil {
		log.Fatal(err)
	}

	tradingMarkets := make([]string, 0)
	for _, market := range markets {
		if market.Status == "trading" {
			tradingMarkets = append(tradingMarkets, market.Market)
		}
	}

	ws, err := bitvavo.NewWsClient()
	if err != nil {
		log.Fatal(err)
	}

	// subscribe to all available 'trading' markets
	tickerchn, err := ws.Ticker().Subscribe(tradingMarkets)
	if err != nil {
		log.Fatal(err)
	}

	for tickerEvent := range tickerchn {
		log.Println(tickerEvent)
	}
}
