package http

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/larscom/go-bitvavo/v2/types"
	"github.com/larscom/go-bitvavo/v2/util"
)

const (
	bitvavoURL          = "https://api.bitvavo.com/v2"
	maxWindowTimeMs     = 60000
	defaultWindowTimeMs = 10000

	headerRatelimit        = "Bitvavo-Ratelimit-Remaining"
	headerRatelimitResetAt = "Bitvavo-Ratelimit-Resetat"
	headerAccessKey        = "Bitvavo-Access-Key"
	headerAccessSignature  = "Bitvavo-Access-Signature"
	headerAccessTimestamp  = "Bitvavo-Access-Timestamp"
	headerAccessWindow     = "Bitvavo-Access-Window"
)

type HttpClient interface {
	// GetRateLimit returns the remaining rate limit.
	//
	// Default value: -1
	GetRateLimit() int64

	// GetRateLimitResetAt returns the time (local time) when the counter resets.
	GetRateLimitResetAt() time.Time

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
	GetTimeWithContext(ctx context.Context) (int64, error)

	// GetMarkets returns the available markets with their status (trading,halted,auction) and
	// available order types.
	GetMarkets() ([]types.Market, error)
	GetMarketsWithContext(ctx context.Context) ([]types.Market, error)

	// GetMarkets returns the available markets with their status (trading,halted,auction) and
	// available order types for a single market (e.g: ETH-EUR)
	GetMarket(market string) (types.Market, error)
	GetMarketWithContext(ctx context.Context, market string) (types.Market, error)

	// GetAssets returns information on the supported assets
	GetAssets() ([]types.Asset, error)
	GetAssetsWithContext(ctx context.Context) ([]types.Asset, error)

	// GetAsset returns information on the supported asset by symbol (e.g: ETH).
	GetAsset(symbol string) (types.Asset, error)
	GetAssetWithContext(ctx context.Context, symbol string) (types.Asset, error)

	// GetOrderBook returns a book with bids and asks for market.
	// That is, the buy and sell orders made by all Bitvavo users in a specific market (e.g: ETH-EUR).
	// The orders in the return parameters are sorted by price
	//
	// Optionally provide the depth (single value) to return the top depth orders only.
	GetOrderBook(market string, depth ...uint64) (types.Book, error)
	GetOrderBookWithContext(ctx context.Context, market string, depth ...uint64) (types.Book, error)

	// GetTrades returns the list of all trades made by all Bitvavo users for market (e.g: ETH-EUR).
	// That is, the trades that have been executed in the past.
	//
	// Optionally provide extra params (see: TradeParams)
	GetTrades(market string, params ...OptionalParams) ([]types.Trade, error)
	GetTradesWithContext(ctx context.Context, market string, params ...OptionalParams) ([]types.Trade, error)

	// GetCandles returns the Open, High, Low, Close, Volume (OHLCV) data you use to create candlestick charts
	// for market with interval time between each candlestick (e.g: market=ETH-EUR interval=5m)
	//
	// Optionally provide extra params (see: CandleParams)
	GetCandles(market string, interval string, params ...OptionalParams) ([]types.Candle, error)
	GetCandlesWithContext(ctx context.Context, market string, interval string, params ...OptionalParams) ([]types.Candle, error)

	// GetTickerPrices returns price of the latest trades on Bitvavo for all markets.
	GetTickerPrices() ([]types.TickerPrice, error)
	GetTickerPricesWithContext(ctx context.Context) ([]types.TickerPrice, error)

	// GetTickerPrice returns price of the latest trades on Bitvavo for a single market (e.g: ETH-EUR).
	GetTickerPrice(market string) (types.TickerPrice, error)
	GetTickerPriceWithContext(ctx context.Context, market string) (types.TickerPrice, error)

	// GetTickerBooks returns the highest buy and the lowest sell prices currently available for
	// all markets in the Bitvavo order book.
	GetTickerBooks() ([]types.TickerBook, error)
	GetTickerBooksWithContext(ctx context.Context) ([]types.TickerBook, error)

	// GetTickerBook returns the highest buy and the lowest sell prices currently
	// available for a single market (e.g: ETH-EUR) in the Bitvavo order book.
	GetTickerBook(market string) (types.TickerBook, error)
	GetTickerBookWithContext(ctx context.Context, market string) (types.TickerBook, error)

	// GetTickers24h returns high, low, open, last, and volume information for trades and orders for all markets over the previous 24 hours.
	GetTickers24h() ([]types.Ticker24h, error)
	GetTickers24hWithContext(ctx context.Context) ([]types.Ticker24h, error)

	// GetTicker24h returns high, low, open, last, and volume information for trades and orders for a single market over the previous 24 hours.
	GetTicker24h(market string) (types.Ticker24h, error)
	GetTicker24hWithContext(ctx context.Context, market string) (types.Ticker24h, error)
}

type httpClient struct {
	mu               sync.RWMutex
	ratelimit        int64
	ratelimitResetAt time.Time

	authClient *httpClientAuth
}

func NewHttpClient() HttpClient {
	client := &httpClient{
		ratelimit: -1,
	}

	return client
}

func (c *httpClient) ToAuthClient(apiKey string, apiSecret string, windowTimeMs ...uint64) HttpClientAuth {
	if c.hasAuthClient() {
		return c.authClient
	}

	windowTime := util.IfOrElse(len(windowTimeMs) > 0, func() uint64 { return windowTimeMs[0] }, 0)
	if windowTime == 0 {
		windowTime = defaultWindowTimeMs
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
	return c.GetTimeWithContext(context.Background())
}

func (c *httpClient) GetTimeWithContext(ctx context.Context) (int64, error) {
	resp, err := httpGet[map[string]float64](
		ctx,
		fmt.Sprintf("%s/time", bitvavoURL),
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
	return c.GetMarketsWithContext(context.Background())
}

func (c *httpClient) GetMarketsWithContext(ctx context.Context) ([]types.Market, error) {
	return httpGet[[]types.Market](
		ctx,
		fmt.Sprintf("%s/markets", bitvavoURL),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetMarket(market string) (types.Market, error) {
	return c.GetMarketWithContext(context.Background(), market)
}

func (c *httpClient) GetMarketWithContext(ctx context.Context, market string) (types.Market, error) {
	params := make(url.Values)
	params.Add("market", market)

	return httpGet[types.Market](
		ctx,
		fmt.Sprintf("%s/markets", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetAssets() ([]types.Asset, error) {
	return c.GetAssetsWithContext(context.Background())
}

func (c *httpClient) GetAssetsWithContext(ctx context.Context) ([]types.Asset, error) {
	return httpGet[[]types.Asset](
		ctx,
		fmt.Sprintf("%s/assets", bitvavoURL),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetAsset(symbol string) (types.Asset, error) {
	return c.GetAssetWithContext(context.Background(), symbol)
}

func (c *httpClient) GetAssetWithContext(ctx context.Context, symbol string) (types.Asset, error) {
	params := make(url.Values)
	params.Add("symbol", symbol)

	return httpGet[types.Asset](
		ctx,
		fmt.Sprintf("%s/assets", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetOrderBook(market string, depth ...uint64) (types.Book, error) {
	return c.GetOrderBookWithContext(context.Background(), market, depth...)
}

func (c *httpClient) GetOrderBookWithContext(ctx context.Context, market string, depth ...uint64) (types.Book, error) {
	params := make(url.Values)
	if len(depth) > 0 {
		params.Add("depth", fmt.Sprint(depth[0]))
	}

	return httpGet[types.Book](
		ctx,
		fmt.Sprintf("%s/%s/book", bitvavoURL, market),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTrades(market string, opt ...OptionalParams) ([]types.Trade, error) {
	return c.GetTradesWithContext(context.Background(), market, opt...)
}

func (c *httpClient) GetTradesWithContext(ctx context.Context, market string, opt ...OptionalParams) ([]types.Trade, error) {
	params := make(url.Values)
	if len(opt) > 0 {
		params = opt[0].Params()
	}
	return httpGet[[]types.Trade](
		ctx,
		fmt.Sprintf("%s/%s/trades", bitvavoURL, market),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetCandles(market string, interval string, opt ...OptionalParams) ([]types.Candle, error) {
	return c.GetCandlesWithContext(context.Background(), market, interval, opt...)
}

func (c *httpClient) GetCandlesWithContext(ctx context.Context, market string, interval string, opt ...OptionalParams) ([]types.Candle, error) {
	params := make(url.Values)
	if len(opt) > 0 {
		params = opt[0].Params()
	}
	params.Add("interval", interval)

	return httpGet[[]types.Candle](
		ctx,
		fmt.Sprintf("%s/%s/candles", bitvavoURL, market),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTickerPrices() ([]types.TickerPrice, error) {
	return c.GetTickerPricesWithContext(context.Background())
}

func (c *httpClient) GetTickerPricesWithContext(ctx context.Context) ([]types.TickerPrice, error) {
	return httpGet[[]types.TickerPrice](
		ctx,
		fmt.Sprintf("%s/ticker/price", bitvavoURL),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTickerPrice(market string) (types.TickerPrice, error) {
	return c.GetTickerPriceWithContext(context.Background(), market)
}

func (c *httpClient) GetTickerPriceWithContext(ctx context.Context, market string) (types.TickerPrice, error) {
	params := make(url.Values)
	params.Add("market", market)

	return httpGet[types.TickerPrice](
		ctx,
		fmt.Sprintf("%s/ticker/price", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTickerBooks() ([]types.TickerBook, error) {
	return c.GetTickerBooksWithContext(context.Background())
}

func (c *httpClient) GetTickerBooksWithContext(ctx context.Context) ([]types.TickerBook, error) {
	return httpGet[[]types.TickerBook](
		ctx,
		fmt.Sprintf("%s/ticker/book", bitvavoURL),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTickerBook(market string) (types.TickerBook, error) {
	return c.GetTickerBookWithContext(context.Background(), market)
}

func (c *httpClient) GetTickerBookWithContext(ctx context.Context, market string) (types.TickerBook, error) {
	params := make(url.Values)
	params.Add("market", market)

	return httpGet[types.TickerBook](
		ctx,
		fmt.Sprintf("%s/ticker/book", bitvavoURL),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTickers24h() ([]types.Ticker24h, error) {
	return c.GetTickers24hWithContext(context.Background())
}

func (c *httpClient) GetTickers24hWithContext(ctx context.Context) ([]types.Ticker24h, error) {
	return httpGet[[]types.Ticker24h](
		ctx,
		fmt.Sprintf("%s/ticker/24h", bitvavoURL),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		nil,
	)
}

func (c *httpClient) GetTicker24h(market string) (types.Ticker24h, error) {
	return c.GetTicker24hWithContext(context.Background(), market)
}

func (c *httpClient) GetTicker24hWithContext(ctx context.Context, market string) (types.Ticker24h, error) {
	params := make(url.Values)
	params.Add("market", market)

	return httpGet[types.Ticker24h](
		ctx,
		fmt.Sprintf("%s/ticker/24h", bitvavoURL),
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
