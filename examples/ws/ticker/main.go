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
		tickerchn, err := ws.Ticker().Subscribe("ETH-EUR")
		if err != nil {
			log.Fatal(err)
		}

		for tickerEvent := range tickerchn {
			log.Println(tickerEvent)
		}
	}()
	go func() {
		tickerchn, err := ws.Ticker().Subscribe("1INCH-EUR")
		if err != nil {
			log.Fatal(err)
		}

		for tickerEvent := range tickerchn {
			log.Println(tickerEvent)
		}
	}()
	go func() {
		tickerchn, err := ws.Ticker().Subscribe("AAVE-EUR")
		if err != nil {
			log.Fatal(err)
		}

		for tickerEvent := range tickerchn {
			log.Println(tickerEvent)
		}
	}()
	go func() {
		tickerchn, err := ws.Ticker().Subscribe("ACH-EUR")
		if err != nil {
			log.Fatal(err)
		}

		for tickerEvent := range tickerchn {
			log.Println(tickerEvent)
		}
	}()
	go func() {
		tickerchn, err := ws.Ticker().Subscribe("ADA-EUR")
		if err != nil {
			log.Fatal(err)
		}

		for tickerEvent := range tickerchn {
			log.Println(tickerEvent)
		}
	}()
	go func() {
		tickerchn, err := ws.Ticker().Subscribe("ADX-EUR")
		if err != nil {
			log.Fatal(err)
		}

		for tickerEvent := range tickerchn {
			log.Println(tickerEvent)
		}
	}()
	go func() {
		tickerchn, err := ws.Ticker().Subscribe("AE-EUR")
		if err != nil {
			log.Fatal(err)
		}

		for tickerEvent := range tickerchn {
			log.Println(tickerEvent)
		}
	}()
	go func() {
		tickerchn, err := ws.Ticker().Subscribe("AGIX-EUR")
		if err != nil {
			log.Fatal(err)
		}

		for tickerEvent := range tickerchn {
			log.Println(tickerEvent)
		}
	}()
	go func() {
		tickerchn, err := ws.Ticker().Subscribe("AION-EUR")
		if err != nil {
			log.Fatal(err)
		}

		for tickerEvent := range tickerchn {
			log.Println(tickerEvent)
		}
	}()
	go func() {
		tickerchn, err := ws.Ticker().Subscribe("AKRO-EUR")
		if err != nil {
			log.Fatal(err)
		}

		for tickerEvent := range tickerchn {
			log.Println(tickerEvent)
		}
	}()

	tickerchn, err := ws.Ticker().Subscribe("ALGO-EUR")
	if err != nil {
		log.Fatal(err)
	}

	for tickerEvent := range tickerchn {
		log.Println(tickerEvent)
	}
}
