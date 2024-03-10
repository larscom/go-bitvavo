package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/larscom/go-bitvavo/v2"
)

func main() {
	bitvavo.EnableDebugLogging()

	if err := godotenv.Load(); err != nil {
		log.Println("Starting without .env file")
	}
	var (
		key        = os.Getenv("API_KEY")
		secret     = os.Getenv("API_SECRET")
		client     = bitvavo.NewHttpClient()
		authClient = client.ToAuthClient(key, secret)
	)

	history, err := authClient.GetWithdrawalHistory()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(history)

}
