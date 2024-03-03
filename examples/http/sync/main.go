package main

import (
	"log"

	"github.com/larscom/go-bitvavo/v2"
	"github.com/larscom/go-bitvavo/v2/httpc"
)

func main() {
	client := bitvavo.NewHttpClient(httpc.WithDebug())

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
