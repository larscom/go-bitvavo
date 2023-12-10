package httpc

import (
	"fmt"
	"time"

	"net/url"

	"github.com/larscom/go-bitvavo/v2/types"
)

type HttpClientAuth interface {
	// GetBalance returns the balance on the account.
	// Optionally provide the symbol to filter for in uppercase (e.g: ETH)
	GetBalance(symbol ...string) ([]types.Balance, error)

	// GetAccount returns trading volume and fees for account.
	GetAccount() (types.Account, error)

	// GetOrders returns data for multiple orders at once for market (e.g: ETH-EUR)
	//
	// Optionally provide extra params (see: OrderParams)
	GetOrders(market string, params ...OptionalParams) ([]types.Order, error)

	// GetOrdersOpen returns all open orders for market (e.g: ETH-EUR) or all open orders
	// if no market is given.
	GetOrdersOpen(market ...string) ([]types.Order, error)

	// GetOrder returns the order by market and ID
	GetOrder(market string, orderId string) (types.Order, error)

	// CancelOrders cancels multiple orders at once.
	// Either for an entire market (e.g: ETH-EUR) or for the entire account if you
	// omit the market.
	//
	// It returns a slice of orderId's of which are canceled
	CancelOrders(market ...string) ([]string, error)

	// CancelOrder cancels a single order by ID for the specific market (e.g: ETH-EUR)
	//
	// It returns the canceled orderId if it was canceled
	CancelOrder(market string, orderId string) (string, error)

	// CreateOrder places a new order on the exchange.
	//
	// It returns the created order if it was succesfully created
	CreateOrder(market string, side string, orderType string, order types.OrderCreate) (types.Order, error)
}

type httpClientAuth struct {
	config                 *authConfig
	updateRateLimit        func(ratelimit int64)
	updateRateLimitResetAt func(resetAt time.Time)
	logDebug               func(message string, args ...any)
}

type authConfig struct {
	apiKey       string
	apiSecret    string
	windowTimeMs uint64
}

func newHttpClientAuth(
	updateRateLimit func(ratelimit int64),
	updateRateLimitResetAt func(resetAt time.Time),
	logDebug func(message string, args ...any),
	config *authConfig,
) *httpClientAuth {
	return &httpClientAuth{
		updateRateLimit:        updateRateLimit,
		updateRateLimitResetAt: updateRateLimitResetAt,
		logDebug:               logDebug,
		config:                 config,
	}
}

func (c *httpClientAuth) GetBalance(symbol ...string) ([]types.Balance, error) {
	params := make(url.Values)
	if len(symbol) > 0 {
		params.Add("symbol", symbol[0])
	}

	return httpGet[[]types.Balance](
		fmt.Sprintf("%s/balance", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		c.config,
	)
}

func (c *httpClientAuth) GetAccount() (types.Account, error) {
	return httpGet[types.Account](
		fmt.Sprintf("%s/account", httpUrl),
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		c.config,
	)
}

func (c *httpClientAuth) GetOrders(market string, opt ...OptionalParams) ([]types.Order, error) {
	params := make(url.Values)
	if len(opt) > 0 {
		params = opt[0].Params()
	}
	params.Add("market", market)

	return httpGet[[]types.Order](
		fmt.Sprintf("%s/orders", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		c.config,
	)
}

func (c *httpClientAuth) GetOrdersOpen(market ...string) ([]types.Order, error) {
	params := make(url.Values)
	if len(market) > 0 {
		params.Add("market", market[0])
	}

	return httpGet[[]types.Order](
		fmt.Sprintf("%s/ordersOpen", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		c.config,
	)
}

func (c *httpClientAuth) GetOrder(market string, orderId string) (types.Order, error) {
	params := make(url.Values)
	params.Add("market", market)
	params.Add("orderId", orderId)

	return httpGet[types.Order](
		fmt.Sprintf("%s/order", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		c.config,
	)
}

func (c *httpClientAuth) CancelOrders(market ...string) ([]string, error) {
	params := make(url.Values)
	if len(market) > 0 {
		params.Add("market", market[0])
	}

	resp, err := httpDelete[[]map[string]string](
		fmt.Sprintf("%s/orders", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		c.config,
	)
	if err != nil {
		return nil, err
	}

	orderIds := make([]string, len(resp))
	for i := 0; i < len(orderIds); i++ {
		orderIds[i] = resp[i]["orderId"]
	}

	return orderIds, nil
}

func (c *httpClientAuth) CancelOrder(market string, orderId string) (string, error) {
	params := make(url.Values)
	params.Add("market", market)
	params.Add("orderId", orderId)

	resp, err := httpDelete[map[string]string](
		fmt.Sprintf("%s/order", httpUrl),
		params,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		c.config,
	)
	if err != nil {
		return "", err
	}

	return resp["orderId"], nil
}

func (c *httpClientAuth) CreateOrder(market string, side string, orderType string, order types.OrderCreate) (types.Order, error) {
	order.Market = market
	order.Side = side
	order.OrderType = orderType
	return httpPost[types.Order](
		fmt.Sprintf("%s/order", httpUrl),
		order,
		emptyParams,
		c.updateRateLimit,
		c.updateRateLimitResetAt,
		c.logDebug,
		c.config,
	)
}
