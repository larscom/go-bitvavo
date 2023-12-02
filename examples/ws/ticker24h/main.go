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

	ticker24hchn, err := ws.Ticker24h().Subscribe("ETH-EUR", 0)
	if err != nil {
		log.Fatal(err)
	}

	for ticker24hEvent := range ticker24hchn {
		log.Println(ticker24hEvent)
	}
}
