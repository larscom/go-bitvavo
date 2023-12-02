package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/larscom/go-bitvavo/v2"
	"github.com/larscom/go-bitvavo/v2/httpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Starting without .env file")
	}
	var (
		key        = os.Getenv("API_KEY")
		secret     = os.Getenv("API_SECRET")
		client     = bitvavo.NewHttpClient(httpc.WithDebug(false))
		authClient = client.ToAuthClient(key, secret, 0)
	)

	balance, err := authClient.GetBalance()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Balance", balance)

	account, err := authClient.GetAccount()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Account", account)

	ratelimit := client.GetRateLimit()
	resetAt := client.GetRateLimitResetAt()
	log.Println("RateLimit", ratelimit, "ResetAt", resetAt)
}
