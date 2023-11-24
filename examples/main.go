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

	ws, err := bitvavo.NewWebSocket(bitvavo.WithDebug(false))
	if err != nil {
		log.Fatal(err)
	}

	chn, err := ws.Candles().Subscribe("ETH-EUR", "5m")
	if err != nil {
		log.Fatal(err)
	}

	for value := range chn {
		log.Println("value", value)
	}
}
