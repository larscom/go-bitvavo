package main

import (
	"log"

	"github.com/larscom/go-bitvavo/v2"
	"github.com/larscom/go-bitvavo/v2/wsc"
)

func main() {
	ws, err := bitvavo.NewWsClient(wsc.WithDebug())
	if err != nil {
		log.Fatal(err)
	}

	candlechn, err := ws.Candles().Subscribe("BTC-EUR", "5m")
	if err != nil {
		log.Fatal(err)
	}

	for candleEvent := range candlechn {
		log.Println(candleEvent)
	}
}
