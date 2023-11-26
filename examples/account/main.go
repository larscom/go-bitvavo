package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/larscom/go-bitvavo/v2"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Starting without .env file")
	}

	ws, err := bitvavo.NewWebSocket()
	if err != nil {
		log.Fatal(err)
	}

	key := os.Getenv("API_KEY")
	secret := os.Getenv("API_SECRET")

	account, err := ws.Account(key, secret).Subscribe("ETH-EUR")
	if err != nil {
		log.Fatal(err)
	}

	for orderEvent := range account.Order(50) {
		log.Println(orderEvent)
	}
}
