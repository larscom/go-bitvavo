package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/larscom/go-bitvavo/v2"
	"github.com/larscom/go-bitvavo/v2/ws"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Starting without .env file")
	}

	key := os.Getenv("API_KEY")
	secret := os.Getenv("API_SECRET")

	ws, err := bitvavo.NewWsClient(ws.WithDebug())
	if err != nil {
		log.Fatal(err)
	}

	account, err := ws.Account(key, secret).Subscribe("ETH-EUR")
	if err != nil {
		log.Fatal(err)
	}

	for orderEvent := range account.Order(50) {
		log.Println(orderEvent)
	}
}
