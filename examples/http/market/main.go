package main

import (
	"log"

	"github.com/larscom/go-bitvavo/v2"
	"github.com/larscom/go-bitvavo/v2/httpc"
)

func main() {
	client := bitvavo.NewHttpClient(httpc.WithDebug(true))

	book, err := client.GetOrderBook("ETH-EUR")
	if err != nil {
		log.Panic(err)
	}

	log.Println("Book", book)
}
