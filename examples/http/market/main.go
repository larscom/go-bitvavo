package main

import (
	"log"
	"time"

	"github.com/larscom/go-bitvavo/v2"
	"github.com/larscom/go-bitvavo/v2/httpc"
	"github.com/larscom/go-bitvavo/v2/types"
)

func main() {
	client := bitvavo.NewHttpClient(httpc.WithDebug(true))

	book, err := client.GetOrderBook("ETH-EUR")
	if err != nil {
		log.Panic(err)
	}
	log.Println("Book", book)

	trades, err := client.GetTrades("ETH-EUR", &types.TradeParams{
		Start: time.Now().Add(-1 * time.Minute),
	})
	if err != nil {
		log.Panic(err)
	}
	log.Println("Trades", trades)
}
