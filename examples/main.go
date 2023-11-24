package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/larscom/go-bitvavo/v2"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Starting without .env file")
	}

	ws, err := bitvavo.NewWebSocket(bitvavo.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	tickerchn, err := ws.Ticker().Subscribe("ETH-EUR")
	if err != nil {
		log.Fatal(err)
	}

	for ticker := range tickerchn {
		log.Println("ticker", ticker)
	}
}
