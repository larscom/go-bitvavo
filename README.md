# Go Bitvavo

[![Go Report Card](https://goreportcard.com/badge/github.com/larscom/go-bitvavo/v2)](https://goreportcard.com/report/github.com/larscom/go-bitvavo/v2)
[![Go Reference](https://pkg.go.dev/badge/github.com/larscom/go-bitvavo.svg)](https://pkg.go.dev/github.com/larscom/go-bitvavo)

> Go **thread safe** client library (WebSockets / HTTP) for Bitvavo v2 (https://docs.bitvavo.com)

Go Bitvavo is a **thread-safe** client written in GO to interact with the Bitvavo platform. It includes a WebSocket client (for read-only purposes) to listen to all events occurring on the Bitvavo platform (e.g. candles, ticker, orders, fills, etc.) and an HTTP client (for read/write operations). The HTTP client can retrieve the same data as WebSockets but can also perform write operations such as placing orders and withdrawing assets from your account.

## 📒 Supported Features

- WebSocket Client -- Read only
  - Account (orders/fills)
  - Book
  - Candles
  - Trades
  - Ticker
  - Ticker 24h
- Http Client -- Read / Write
  - Market data endpoints
  - Account endpoints
  - Synchronization endpoints
  - Trading endpoints
  - Transfer endpoints

## 🚀 Installation

```shell
go get github.com/larscom/go-bitvavo/v2@latest
```

## 💡 Usage

```shell
import "github.com/larscom/go-bitvavo/v2"
```

## 🌐 HTTP client

Every function on the HTTP client has a `WithContext` variant so you can pass your own context as argument.

```go
client := bitvavo.NewHttpClient()
```

### Public endpoints

```go
func main() {
	client := bitvavo.NewHttpClient()

	time, err := client.GetTime()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(time)

	// provide your own context
	time, err := client.GetTimeWithContext(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	log.Println(time)

}

```

### Private endpoints

```go
func main() {
	client := bitvavo.NewHttpClient()

	// create a new auth client for authenticated requests
	authClient = client.ToAuthClient("MY_API_KEY", "MY_API_SECRET")

	balance, err := authClient.GetBalance("ETH")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Balance", balance)
}

```

## 👂 WebSocket client

By default, the websocket handler will try to reconnect to the websocket when the connection is lost, you can disable this behaviour in the options.

For each subscription you can set the buffer size for the underlying channel. All channels have a default buffer size of `50` which should be
sufficient in most cases. You may need to increase this number if you have a **large** amount of subscriptions.

### Public Subscriptions

Public subscriptions requires no authentication and can be used directly.

#### Candles

Subscribe to candle events for market: `ETH-EUR` with an interval of `5m`

```go
func main() {
	ws, err := bitvavo.NewWsClient()
	if err != nil {
		log.Fatal(err)
	}

	chn, err := ws.Candles().Subscribe([]string{"ETH-EUR"}, "5m")
	if err != nil {
		log.Fatal(err)
	}

	for candlesEvent := range chn {
		log.Println(candlesEvent)
	}
}

```

<details>
 <summary>View Event</summary>

```go
type CandlesEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The interval which was requested in the subscription.
	Interval string `json:"interval"`

	// The candle in the defined time period.
	Candle Candle `json:"candle"`
}
...
```

</details>

#### Book

Subscribe to book events for market: `ETH-EUR`

```go
func main() {
	ws, err := bitvavo.NewWsClient()
	if err != nil {
		log.Fatal(err)
	}

	bookchn, err := ws.Book().Subscribe([]string{"ETH-EUR"})
	if err != nil {
		log.Fatal(err)
	}

	for bookEvent := range bookchn {
		log.Println(bookEvent)
	}
}

```

<details>
 <summary>View Event</summary>

```go
type BookEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The book containing the bids and asks.
	Book Book `json:"book"`
}
...
```

</details>

#### Ticker

Subscribe to ticker events for market: `ETH-EUR`

```go
func main() {
	ws, err := bitvavo.NewWsClient()
	if err != nil {
		log.Fatal(err)
	}

	tickerchn, err := ws.Ticker().Subscribe([]string{"ETH-EUR"})
	if err != nil {
		log.Fatal(err)
	}

	for tickerEvent := range tickerchn {
		log.Println(tickerEvent)
	}
}

```

<details>
 <summary>View Event</summary>

```go
type TickerEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The ticker containing the prices.
	Ticker Ticker `json:"ticker"`
}
...
```

</details>

#### Ticker 24H

Subscribe to ticker24h events for market: `ETH-EUR`

```go
func main() {
	ws, err := bitvavo.NewWsClient()
	if err != nil {
		log.Fatal(err)
	}

	ticker24hchn, err := ws.Ticker24h().Subscribe([]string{"ETH-EUR"})
	if err != nil {
		log.Fatal(err)
	}

	for ticker24hEvent := range ticker24hchn {
		log.Println(ticker24hEvent)
	}
}

```

<details>
 <summary>View Event</summary>

```go
type Ticker24hEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The ticker24h containing the prices etc.
	Ticker24h Ticker24h `json:"ticker24h"`
}
...
```

</details>

#### Trades

Subscribe to trades events for market: `ETH-EUR`

```go
func main() {
	ws, err := bitvavo.NewWsClient()
	if err != nil {
		log.Fatal(err)
	}

	tradeschn, err := ws.Trades().Subscribe([]string{"ETH-EUR"})
	if err != nil {
		log.Fatal(err)
	}

	for tradesEvent := range tradeschn {
		log.Println(tradesEvent)
	}
}

```

<details>
 <summary>View Event</summary>

```go
type TradesEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The trade containing the price, side etc.
	Trade Trade `json:"trade"`
}
...
```

</details>

### Private Subscriptions

Private subscriptions do require authentication in the form of an `API key` and `API secret` which you can setup in Bitvavo.

#### Account :: Orders / Fills

Subscribe to order and fill events for market: `ETH-EUR`

```go
func main() {
	ws, err := bitvavo.NewWsClient()
	if err != nil {
		log.Fatal(err)
	}

	key := "MY API KEY"
	secret := "MY API SECRET"

	orderchn, fillchn, err := ws.Account(key, secret).Subscribe([]string{"ETH-EUR"})
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case orderEvent := <-orderchn:
			log.Println(orderEvent)

		case fillEvent := <-fillchn:
			log.Println(fillEvent)
		}
	}
}

```

<details>
 <summary>View Event(s)</summary>

```go
type OrderEvent struct {
	// Describes the returned event over the socket.
	Event string `json:"event"`

	// The market which was requested in the subscription.
	Market string `json:"market"`

	// The order itself.
	Order Order `json:"order"`
}

type FillEvent struct {
	// Describes the returned event over the socket
	Event string `json:"event"`
	// The market which was requested in the subscription
	Market string `json:"market"`
	// The fill itself
	Fill Fill `json:"fill"`
}
...
```

</details>

## 🔧 Options

### Auto Reconnect

You can disable auto reconnecting to the websocket by providing an option to the Websocket constructor

```go
ws, err := bitvavo.NewWsClient(ws.WithAutoReconnect(false))
```

### Error Channel

You can provide your own error channel if you wish to receive errors (e.g. reconnect error, auth error, write failed, read failed)

```go
errchn := make(chan error)
ws, _ := bitvavo.NewWsClient(ws.WithErrorChannel(errchn))

for err := range errchn {
	log.Println(err)
}
```

## 📝 Debug Logging

You can enable debug logging for the HTTP client and WS client.

```go
bitvavo.EnableDebugLogging()
```
