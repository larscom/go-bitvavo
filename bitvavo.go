package bitvavo

import (
	"github.com/larscom/go-bitvavo/v2/httpc"
	"github.com/larscom/go-bitvavo/v2/wsc"
)

// NewWsClient creates a new Websocket client
func NewWsClient(options ...wsc.Option) (wsc.WsClient, error) {
	return wsc.NewWsClient(options...)
}

// NewHttpClient creates a new HTTP client to make unauthenticated requests.
// For authenticated requests, call ToAuthClient func on this HttpClient
func NewHttpClient(options ...httpc.Option) httpc.HttpClient {
	return httpc.NewHttpClient(options...)
}
