package main

import (
	"log"

	"github.com/larscom/go-bitvavo/v2"
)

func main() {
	client := bitvavo.NewHttpClient()
	markets, err := client.GetMarkets()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Markets", markets)
}
