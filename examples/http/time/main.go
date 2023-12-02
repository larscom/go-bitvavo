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
}
