package httpc

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/larscom/go-bitvavo/v2/log"
	"github.com/larscom/go-bitvavo/v2/types"
	"github.com/larscom/go-bitvavo/v2/util"
)

const (
	httpUrl         = "https://api.bitvavo.com/v2"
	maxWindowTimeMs = 60000

	headerRatelimit        = "Bitvavo-Ratelimit-Remaining"
	headerRatelimitResetAt = "Bitvavo-Ratelimit-Resetat"
	headerAccessKey        = "Bitvavo-Access-Key"
	headerAccessSignature  = "Bitvavo-Access-Signature"
	headerAccessTimestamp  = "Bitvavo-Access-Timestamp"
	headerAccessWindow     = "Bitvavo-Access-Window"
)
const DefaultWindowTimeMs = 10000

type HttpClient interface {
	// ToAuthClient returns a client for authenticated requests.
	// You need to provide an apiKey and an apiSecret which you can create in the bitvavo dashboard.
	//
	// WindowTimeMs is the window that allows execution of your request.
	//
	// If you set the value to 0, the default value of 10000 will be set.
	// Whenever you go higher than the max value of 60000 the value will be set to 60000.
	ToAuthClient(apiKey string, apiSecret string, windowTimeMs ...uint64) HttpClientAuth

	// GetTime returns the current server time in milliseconds since 1 Jan 1970
	GetTime() (int64, error)

	// GetRateLimit returns the remaining rate limit.
	//
	// Default value: -1
	GetRateLimit() int64

	// GetRateLimitResetAt returns the time (local time) when the counter resets.
	GetRateLimitResetAt() time.Time

	// GetMarkets returns the available markets with their status (trading,halted,auction) and
	// available order types.
	GetMarkets() ([]types.Market, error)

	// GetMarkets returns the available markets with their status (trading,halted,auction) and
	// available order types for a single market (e.g: ETH-EUR)
	GetMarket(market string) (types.Market, error)

	// GetAssets returns information on the supported assets
	GetAssets() ([]types.Asset, error)

	// GetAsset returns information on the supported asset by symbol (e.g: ETH).
	GetAsset(symbol string) (types.Asset, error)

	// GetOrderBook returns a book with bids and asks for market.
	// That is, the buy and sell orders made by all Bitvavo users in a specific market (e.g: ETH-EUR).
	// The orders in the return parameters are sorted by price
	//
	// Optionally provide the depth to return the top depth orders only.
	GetOrderBook(market string, depth ...uint64) (types.Book, error)
}

type Option func(*httpClient)

type httpClient struct {
	debug bool

	mu               sync.RWMutex
	ratelimit        int64
	ratelimitResetAt time.Time

	authClient *httpClientAuth
}

func NewHttpClient(options ...Option) HttpClient {
	client := &httpClient{
		ratelimit: -1,
	}

	for _, opt := range options {
		opt(client)
	}

	return client
}

// Enable debug logging.
// default: false
func WithDebug(debug bool) Option {
	return func(c *httpClient) {
		c.debug = debug
	}
}

func (c *httpClient) ToAuthClient(apiKey string, apiSecret string, windowTimeMs ...uint64) HttpClientAuth {
	if c.hasAuthClient() {
		return c.authClient
	}

	windowTime := util.IfOrElse(len(windowTimeMs) > 0, func() uint64 { return windowTimeMs[0] }, 0)
	if windowTime == 0 {
		windowTime = DefaultWindowTimeMs
	}
	if windowTime > maxWindowTimeMs {
		windowTime = maxWindowTimeMs
	}

	config := &authConfig{
		windowTimeMs: windowTime,
		apiKey:       apiKey,
		apiSecret:    apiSecret,
	}

	c.authClient = newHttpClientAuth(c.updateRateLimit, c.updateRateLimitResetAt, c.logDebug, config)
	return c.authClient
}

func (c *httpClient) GetRateLimit() int64 {
	return c.ratelimit
}

func (c *httpClient) GetRateLimitResetAt() time.Time {
	return c.ratelimitResetAt
}

func (c *httpClient) GetTime() (int64, error) {
	resp, err := httpGet[map[string]float64](
		fmt.Sprintf("%s/time", httpUrl),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		nil,
	)
	if err != nil {
		return 0, err
	}

	return int64(resp["time"]), nil
}

func (c *httpClient) GetMarkets() ([]types.Market, error) {
	return httpGet[[]types.Market](
		fmt.Sprintf("%s/markets", httpUrl),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		nil,
	)
}

func (c *httpClient) GetMarket(market string) (types.Market, error) {
	params := make(url.Values)
	params.Add("market", market)

	return httpGet[types.Market](
		fmt.Sprintf("%s/markets", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		nil,
	)
}

func (c *httpClient) GetAssets() ([]types.Asset, error) {
	return httpGet[[]types.Asset](
		fmt.Sprintf("%s/assets", httpUrl),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		nil,
	)
}

func (c *httpClient) GetAsset(symbol string) (types.Asset, error) {
	params := make(url.Values)
	params.Add("symbol", symbol)

	return httpGet[types.Asset](
		fmt.Sprintf("%s/assets", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		nil,
	)
}

func (c *httpClient) GetOrderBook(market string, depth ...uint64) (types.Book, error) {
	params := make(url.Values)
	if len(depth) > 0 {
		params.Add("depth", fmt.Sprint(depth[0]))
	}

	return httpGet[types.Book](
		fmt.Sprintf("%s/%s/book", httpUrl, market),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		nil,
	)
}

func (c *httpClient) updateRateLimit(ratelimit int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ratelimit = ratelimit
}

func (c *httpClient) updateRateLimitResetAt(resetAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ratelimitResetAt = resetAt
}

func (c *httpClient) hasAuthClient() bool {
	return c.authClient != nil
}

func (c *httpClient) logDebug(message string, args ...any) {
	if c.debug {
		log.Logger().Debug(message, args...)
	}
}
