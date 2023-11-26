package main

import (
	"log"

	"github.com/larscom/go-bitvavo/v2"
)

func main() {
	ws, err := bitvavo.NewWebSocket(bitvavo.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	candlechn, err := ws.Candles().Subscribe("ETH-EUR", "5m", 0)
	if err != nil {
		log.Fatal(err)
	}

	for value := range candlechn {
		log.Println("value", value)
	}
}
