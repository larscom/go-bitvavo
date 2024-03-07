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
	tickerchn, err := ws.Ticker().Subscribe([]string{"ETH-EUR",
		"1INCH-EUR", "AAVE-EUR", "ACH-EUR", "ADA-EUR", "ADX-EUR",
		"AE-EUR", "AGIX-EUR", "AION-EUR", "AKRO-EUR",
	})
	if err != nil {
		log.Fatal(err)
	}

	for tickerEvent := range tickerchn {
		log.Println(tickerEvent)
	}
}
