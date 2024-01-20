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

	go func() {
		bookchn, err := ws.Book().Subscribe("ETH-EUR")
		if err != nil {
			log.Fatal(err)
		}

		for bookEvent := range bookchn {
			log.Println(bookEvent)
		}
	}()

	go func() {
		bookchn, err := ws.Book().SubscribeToMarkets([]string{"ETH-EUR", "BTC-EUR"})
		if err != nil {
			log.Fatal(err)
		}

		for bookEvent := range bookchn {
			log.Println(bookEvent)
		}
	}()

	select {}
}
