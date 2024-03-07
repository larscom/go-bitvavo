package main

import (
	"log"

	"github.com/larscom/go-bitvavo/v2"
)

func main() {
	ws, err := bitvavo.NewWsClient()
	if err != nil {
		log.Fatal(err)
	}

	candlechn, err := ws.Candles().Subscribe([]string{"BTC-EUR", "ETH-EUR", "XLM-EUR"}, "5m")
	if err != nil {
		log.Fatal(err)
	}

	for candleEvent := range candlechn {
		log.Println(candleEvent)
	}
}
