package httpc

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sync"
	"time"

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
	// Optionally provide the depth (single value) to return the top depth orders only.
	GetOrderBook(market string, depth ...uint64) (types.Book, error)

	// GetTrades returns the list of all trades made by all Bitvavo users for market (e.g: ETH-EUR).
	// That is, the trades that have been executed in the past.
	//
	// Optionally provide extra params (see: TradeParams)
	GetTrades(market string, params ...OptionalParams) ([]types.Trade, error)

	// GetCandles returns the Open, High, Low, Close, Volume (OHLCV) data you use to create candlestick charts
	// for market with interval time between each candlestick (e.g: market=ETH-EUR interval=5m)
	//
	// Optionally provide extra params (see: CandleParams)
	GetCandles(market string, interval string, params ...OptionalParams) ([]types.Candle, error)

	// GetTickerPrices returns price of the latest trades on Bitvavo for all markets.
	GetTickerPrices() ([]types.TickerPrice, error)

	// GetTickerPrice returns price of the latest trades on Bitvavo for a single market (e.g: ETH-EUR).
	GetTickerPrice(market string) (types.TickerPrice, error)

	// GetTickerBooks returns the highest buy and the lowest sell prices currently available for
	// all markets in the Bitvavo order book.
	GetTickerBooks() ([]types.TickerBook, error)

	// GetTickerBook returns the highest buy and the lowest sell prices currently
	// available for a single market (e.g: ETH-EUR) in the Bitvavo order book.
	GetTickerBook(market string) (types.TickerBook, error)

	// GetTickers24h returns high, low, open, last, and volume information for trades and orders for all markets over the previous 24 hours.
	GetTickers24h() ([]types.Ticker24h, error)

	// GetTicker24h returns high, low, open, last, and volume information for trades and orders for a single market over the previous 24 hours.
	GetTicker24h(market string) (types.Ticker24h, error)
}

type Option func(*httpClient)

type httpClient struct {
	mu               sync.RWMutex
	ratelimit        int64
	ratelimitResetAt time.Time

	authClient *httpClientAuth
}

func NewHttpClient(options ...Option) HttpClient {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	client := &httpClient{
		ratelimit: -1,
	}
	for _, opt := range options {
		opt(client)
	}

	return client
}

// Enable debug logging.
func WithDebug() Option {
	return func(c *httpClient) {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
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

	c.authClient = newHttpClientAuth(c.updateRateLimit, c.updateRateLimitResetAt, config)
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
		nil,
	)
}

func (c *httpClient) GetAssets() ([]types.Asset, error) {
	return httpGet[[]types.Asset](
		fmt.Sprintf("%s/assets", httpUrl),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
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
		nil,
	)
}

func (c *httpClient) GetTrades(market string, opt ...OptionalParams) ([]types.Trade, error) {
	params := make(url.Values)
	if len(opt) > 0 {
		params = opt[0].Params()
	}
	return httpGet[[]types.Trade](
		fmt.Sprintf("%s/%s/trades", httpUrl, market),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetCandles(market string, interval string, opt ...OptionalParams) ([]types.Candle, error) {
	params := make(url.Values)
	if len(opt) > 0 {
		params = opt[0].Params()
	}
	params.Add("interval", interval)

	return httpGet[[]types.Candle](
		fmt.Sprintf("%s/%s/candles", httpUrl, market),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTickerPrices() ([]types.TickerPrice, error) {
	return httpGet[[]types.TickerPrice](
		fmt.Sprintf("%s/ticker/price", httpUrl),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTickerPrice(market string) (types.TickerPrice, error) {
	params := make(url.Values)
	params.Add("market", market)

	return httpGet[types.TickerPrice](
		fmt.Sprintf("%s/ticker/price", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTickerBooks() ([]types.TickerBook, error) {
	return httpGet[[]types.TickerBook](
		fmt.Sprintf("%s/ticker/book", httpUrl),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTickerBook(market string) (types.TickerBook, error) {
	params := make(url.Values)
	params.Add("market", market)

	return httpGet[types.TickerBook](
		fmt.Sprintf("%s/ticker/book", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTickers24h() ([]types.Ticker24h, error) {
	return httpGet[[]types.Ticker24h](
		fmt.Sprintf("%s/ticker/24h", httpUrl),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTicker24h(market string) (types.Ticker24h, error) {
	params := make(url.Values)
	params.Add("market", market)

	return httpGet[types.Ticker24h](
		fmt.Sprintf("%s/ticker/24h", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
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
