package main

import (
	"log"

	"github.com/larscom/go-bitvavo/v2"
)

func main() {
	client := bitvavo.NewHttpClient()

	time, err := client.GetTime()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Time", time)

	markets, err := client.GetMarkets()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Markets", markets)

	assets, err := client.GetAssets()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Assets", assets)
}
