package main

import (
	"log"
	"os"
	"runtime/trace"
	"time"

	"github.com/larscom/go-bitvavo/v2"
)

func main() {
	file := createTraceFile("trace.out")
	defer file.Close()

	if err := trace.Start(file); err != nil {
		log.Fatal(err)
	}
	defer trace.Stop()

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

	go func() {
		time.Sleep(time.Second * 10)
		ws.Ticker().UnsubscribeAll()
	}()

	// subscribe to all available 'trading' markets
	tickerchn, err := ws.Ticker().Subscribe(tradingMarkets)
	if err != nil {
		log.Fatal(err)
	}

	for tickerEvent := range tickerchn {
		log.Println(tickerEvent)
	}
}

func createTraceFile(name string) *os.File {
	file, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	return file
}
