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

	tradeschn, err := ws.Trades().Subscribe("ETH-EUR", 0)
	if err != nil {
		log.Fatal(err)
	}

	for tradesEvent := range tradeschn {
		log.Println(tradesEvent)
	}
}
