package main

import (
	"log"

	"github.com/larscom/go-bitvavo/v2"
)

func main() {
	ws, err := bitvavo.NewWebSocket()
	if err != nil {
		log.Fatal(err)
	}

	bookchn, err := ws.Book().Subscribe("ETH-EUR", 0)
	if err != nil {
		log.Fatal(err)
	}

	for bookEvent := range bookchn {
		log.Println(bookEvent)
	}
}
