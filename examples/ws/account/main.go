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

	key := os.Getenv("API_KEY")
	secret := os.Getenv("API_SECRET")

	ws, err := bitvavo.NewWsClient()
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	orderchn, fillchn, err := ws.Account(key, secret).Subscribe([]string{"ETH-EUR", "BTC-EUR"})
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case orderEvent := <-orderchn:
			log.Println(orderEvent)

		case fillEvent := <-fillchn:
			log.Println(fillEvent)
		}
	}
}
