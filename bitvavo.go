package bitvavo

import (
	"github.com/larscom/go-bitvavo/v2/http"
	"github.com/larscom/go-bitvavo/v2/ws"
	"github.com/rs/zerolog"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
}

// Enable debug logging for the WsClient and HttpClient
func EnableDebugLogging() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

// NewWsClient creates a new Bitvavo Websocket client
func NewWsClient(options ...ws.Option) (ws.WsClient, error) {
	return ws.NewWsClient(options...)
}

// NewHttpClient creates a new Bitvavo HTTP client to make unauthenticated requests.
//
// For authenticated requests, call ToAuthClient func on this HttpClient
func NewHttpClient() http.HttpClient {
	return http.NewHttpClient()
}
